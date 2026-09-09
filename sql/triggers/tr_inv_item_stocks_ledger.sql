-- Guarantees a row in dbo.tbl_inv_stock_transactions for every INSERT/UPDATE/DELETE
-- against dbo.tbl_inv_item_stocks, regardless of which application code path did it
-- (the shared item_stock_service.go functions, the raw reversal blocks in
-- receiving_report_service.go / pick_activity_service2.go, or a future path that
-- forgets to call any of them). qty_before/qty_after/qty_change/direction come
-- straight from the inserted/deleted pseudo-tables, so they can never drift from what
-- actually happened to the row.
--
-- source_type/source_id/remarks/unit_cost/supplier_id/supplier/purchase_date are all
-- optional enrichment: the app sets them right before a write via
-- services.SetStockAuditContext(tx, ...), which writes one row to
-- tbl_inv_stock_audit_context keyed by @@SPID (it used sp_set_session_context until
-- 2026-09-08 - that is SQL Server 2016+, and production is 2012). If a
-- write doesn't set that context, the ledger row is still created - just with those
-- columns NULL - so the numeric trail is never lost, only the "why"/"at what cost".
-- The cost columns are only ever set on writes that went through FIFO lot consumption
-- (see ConsumeLotsFIFO/CreateStockLot in item_stock_services).
--
-- Stub-then-ALTER so this is idempotent on every app startup (see
-- migrations.RunSQLMigrations): the guard below creates a placeholder only when the
-- trigger is absent, and the ALTER then installs the real body. CREATE OR ALTER would
-- read better but does not exist before SQL Server 2016 SP1, and the target server is
-- 2012 - it would be a syntax error there, not a fallback.
IF NOT EXISTS (SELECT 1 FROM sys.triggers WHERE name = 'tr_inv_item_stocks_ledger' AND parent_id = OBJECT_ID('[dbo].[tbl_inv_item_stocks]'))
BEGIN
    EXEC('CREATE TRIGGER [dbo].[tr_inv_item_stocks_ledger] ON [dbo].[tbl_inv_item_stocks] AFTER INSERT AS SET NOCOUNT ON;')
END
GO
ALTER TRIGGER [dbo].[tr_inv_item_stocks_ledger]
ON [dbo].[tbl_inv_item_stocks]
AFTER INSERT, UPDATE, DELETE
AS
BEGIN
    SET NOCOUNT ON;

    -- Read from tbl_inv_stock_audit_context, not SESSION_CONTEXT.
    --
    -- SESSION_CONTEXT() and sp_set_session_context are SQL Server 2016 and later.
    -- On the 2012 production server the DECLARE lines below called an
    -- unrecognised function, so the batch failed to compile and every reference
    -- cascaded into "Must declare the scalar variable @srcType" - the trigger
    -- never installed, and RunSQLMigrations stopped there.
    --
    -- A database's compatibility level does NOT gate that syntax. On a 2016+
    -- engine SESSION_CONTEXT runs fine at level 110, which is exactly why this
    -- passed on the dev box and failed on the server.
    --
    -- CONTEXT_INFO() is the usual 2012 substitute but holds 128 bytes; these
    -- values reach about a thousand characters. Hence a table, keyed by @@SPID:
    -- the write and the trigger it fires are always the same session, so this
    -- reads back exactly what services.SetStockAuditContext just wrote.
    --
    -- Still optional. No row means no context was set, every column below comes
    -- back NULL, and the ledger row is written regardless - qty_before/qty_after/
    -- qty_change/direction come from inserted/deleted, never from here.
    DECLARE @srcType NVARCHAR(100);
    DECLARE @srcId NVARCHAR(50);
    DECLARE @remarks NVARCHAR(500);
    DECLARE @unitCost NVARCHAR(50);
    DECLARE @supplierId NVARCHAR(50);
    DECLARE @supplier NVARCHAR(200);
    DECLARE @purchaseDate NVARCHAR(50);

    SELECT
        @srcType      = NULLIF(source_type, ''),
        @srcId        = NULLIF(source_id, ''),
        @remarks      = NULLIF(remarks, ''),
        @unitCost     = NULLIF(unit_cost, ''),
        @supplierId   = NULLIF(supplier_id, ''),
        @supplier     = NULLIF(supplier, ''),
        @purchaseDate = NULLIF(purchase_date, '')
    FROM dbo.tbl_inv_stock_audit_context
    WHERE spid = @@SPID;

    DECLARE @now DATETIME2 = SYSDATETIME();
    DECLARE @dbUser SYSNAME = SUSER_SNAME();

    -- Pure insert: a brand-new item+warehouse+bin row (not also present in deleted).
    INSERT INTO dbo.tbl_inv_stock_transactions
        (ref_id, item_id, warehouse_id, bin_location, stock_uom, doc_no,
         direction, qty_before, qty_after, qty_change,
         source_type, source_id, remarks,
         unit_cost, supplier_id, supplier, purchase_date,
         transaction_at, db_user)
    SELECT
        i.id, i.item_id, i.warehouse_id, i.bin_location, i.stock_uom, i.doc_no,
        'IN', 0, ISNULL(i.stock_qty, 0), ISNULL(i.stock_qty, 0),
        @srcType, TRY_CAST(@srcId AS INT), @remarks,
        TRY_CAST(@unitCost AS DECIMAL(18,4)), TRY_CAST(@supplierId AS INT), @supplier, @purchaseDate,
        @now, @dbUser
    FROM inserted i
    LEFT JOIN deleted d ON d.id = i.id
    WHERE d.id IS NULL
      AND ISNULL(i.stock_qty, 0) <> 0;

    -- Update: row exists in both inserted and deleted - only log if qty actually moved.
    INSERT INTO dbo.tbl_inv_stock_transactions
        (ref_id, item_id, warehouse_id, bin_location, stock_uom, doc_no,
         direction, qty_before, qty_after, qty_change,
         source_type, source_id, remarks,
         unit_cost, supplier_id, supplier, purchase_date,
         transaction_at, db_user)
    SELECT
        i.id, i.item_id, i.warehouse_id, i.bin_location, i.stock_uom, i.doc_no,
        CASE WHEN ISNULL(i.stock_qty, 0) >= ISNULL(d.stock_qty, 0) THEN 'IN' ELSE 'OUT' END,
        ISNULL(d.stock_qty, 0), ISNULL(i.stock_qty, 0), ISNULL(i.stock_qty, 0) - ISNULL(d.stock_qty, 0),
        @srcType, TRY_CAST(@srcId AS INT), @remarks,
        TRY_CAST(@unitCost AS DECIMAL(18,4)), TRY_CAST(@supplierId AS INT), @supplier, @purchaseDate,
        @now, @dbUser
    FROM inserted i
    INNER JOIN deleted d ON d.id = i.id
    WHERE ISNULL(i.stock_qty, 0) <> ISNULL(d.stock_qty, 0);

    -- Delete: bin row removed while it still held stock. No current code path deletes
    -- tbl_inv_item_stocks rows - this is future-proofing so a stock loss can't happen
    -- silently if that ever changes.
    INSERT INTO dbo.tbl_inv_stock_transactions
        (ref_id, item_id, warehouse_id, bin_location, stock_uom, doc_no,
         direction, qty_before, qty_after, qty_change,
         source_type, source_id, remarks,
         unit_cost, supplier_id, supplier, purchase_date,
         transaction_at, db_user)
    SELECT
        d.id, d.item_id, d.warehouse_id, d.bin_location, d.stock_uom, d.doc_no,
        'OUT', ISNULL(d.stock_qty, 0), 0, -ISNULL(d.stock_qty, 0),
        @srcType, TRY_CAST(@srcId AS INT), @remarks,
        TRY_CAST(@unitCost AS DECIMAL(18,4)), TRY_CAST(@supplierId AS INT), @supplier, @purchaseDate,
        @now, @dbUser
    FROM deleted d
    LEFT JOIN inserted i ON i.id = d.id
    WHERE i.id IS NULL
      AND ISNULL(d.stock_qty, 0) <> 0;
END;
