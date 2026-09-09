-- ============================================================================
-- Document-number fixes for the SQL Server 2012 server.
--
-- MANUAL SCRIPT. Nothing in Go runs sql/snippets - see migrations/run_sql.go,
-- which only walks sql/views, sql/procedures and sql/triggers.
--
-- Run this INSTEAD OF expecting the API's migration to handle it. AutoMigrate
-- cannot do either of these jobs:
--   * Part A is a DATA correction. Migration never touches row values.
--   * Part B changes an existing column's TYPE. GORM will not convert a
--     populated column, and this one is additionally tagged `-:migration`
--     on models.ItemRelease so AutoMigrate skips it entirely.
--
-- Both parts are idempotent - running twice changes nothing the second time -
-- and use only syntax available in SQL Server 2012.
-- ============================================================================

SET NOCOUNT ON;
GO

-- ----------------------------------------------------------------------------
-- PART A - pad tbl_purchasing_purchase_order.doc_no to four digits (§2.5)
--
-- doc_no is text here, and frm_PurchaseOrder.DocNoGenerator() padded on its
-- increment path but returned a bare "1" from its three fallbacks. The first
-- purchase order on a database therefore stored "1" while every later one
-- stored "0002", "0003", ... The generator is fixed as of 2026-09-08; this
-- corrects the rows already written.
--
-- Only all-digit values shorter than four characters are touched, so anything
-- already padded, empty, or non-numeric is left exactly as it is.
-- ----------------------------------------------------------------------------
PRINT 'Part A: padding purchase order doc_no';

UPDATE dbo.tbl_purchasing_purchase_order
SET    doc_no = RIGHT('0000' + doc_no, 4)
WHERE  doc_no IS NOT NULL
  AND  doc_no <> ''
  AND  doc_no NOT LIKE '%[^0-9]%'
  AND  LEN(doc_no) < 4;

PRINT '  rows padded: ' + CAST(@@ROWCOUNT AS varchar(10));
GO

-- ----------------------------------------------------------------------------
-- PART B - tbl_inv_item_release.doc_no  nvarchar -> bigint
--
-- Why this is not cosmetic: utils.NextDocNo issues
--     SELECT COALESCE(MAX(doc_no), 0) FROM tbl_inv_item_release
-- and MAX() over a text column is a STRING comparison. Once a "10" exists,
-- MAX returns "9", so the next number computed is 10 again. There is a UNIQUE
-- index on the column, so the eleventh Item Release does not silently duplicate
-- - it fails outright, and no further Item Release can be created.
--
-- bigint also matches the Go model (models.ItemRelease.DocNo is an int) and the
-- audit table z_tbl_inv_item_release_at.doc_no, which is already bigint.
--
-- The unique index has to come off first: SQL Server will not alter the type of
-- an indexed column. It is recreated identically afterwards.
-- ----------------------------------------------------------------------------
PRINT 'Part B: tbl_inv_item_release.doc_no -> bigint';
GO

-- Refuse rather than corrupt, if anything non-numeric ever got in.
IF EXISTS (SELECT 1 FROM dbo.tbl_inv_item_release
           WHERE doc_no IS NOT NULL AND doc_no LIKE '%[^0-9]%')
BEGIN
    RAISERROR('ABORTED: tbl_inv_item_release.doc_no holds non-numeric values; inspect them before converting.', 16, 1);
END
GO

IF EXISTS (SELECT 1 FROM sys.columns c
           JOIN sys.types t ON t.user_type_id = c.user_type_id
           WHERE c.object_id = OBJECT_ID('dbo.tbl_inv_item_release')
             AND c.name = 'doc_no'
             AND t.name <> 'bigint')
BEGIN
    IF EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'uni_tbl_inv_item_release_doc_no'
                 AND object_id = OBJECT_ID('dbo.tbl_inv_item_release'))
    BEGIN
        PRINT '  dropping unique index';
        DROP INDEX uni_tbl_inv_item_release_doc_no ON dbo.tbl_inv_item_release;
    END

    PRINT '  altering column';
    ALTER TABLE dbo.tbl_inv_item_release ALTER COLUMN doc_no bigint NULL;
END
ELSE
BEGIN
    PRINT '  column already bigint - nothing to do';
END
GO

-- Recreated whether or not this run altered anything, so an interrupted earlier
-- run cannot leave the table without its uniqueness guarantee.
IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'uni_tbl_inv_item_release_doc_no'
                 AND object_id = OBJECT_ID('dbo.tbl_inv_item_release'))
BEGIN
    PRINT '  recreating unique index';
    CREATE UNIQUE NONCLUSTERED INDEX uni_tbl_inv_item_release_doc_no
        ON dbo.tbl_inv_item_release (doc_no);
END
GO

-- ----------------------------------------------------------------------------
-- Verification
-- ----------------------------------------------------------------------------
PRINT '';
PRINT 'RESULT';

SELECT  'purchase_order doc_no still unpadded' AS check_name,
        CAST(COUNT(*) AS varchar(10)) AS value
FROM    dbo.tbl_purchasing_purchase_order
WHERE   doc_no IS NOT NULL AND doc_no <> ''
  AND   doc_no NOT LIKE '%[^0-9]%' AND LEN(doc_no) < 4
UNION ALL
SELECT  'item_release doc_no type',
        t.name
FROM    sys.columns c
JOIN    sys.types t ON t.user_type_id = c.user_type_id
WHERE   c.object_id = OBJECT_ID('dbo.tbl_inv_item_release') AND c.name = 'doc_no'
UNION ALL
SELECT  'item_release unique index present',
        CASE WHEN EXISTS (SELECT 1 FROM sys.indexes
                          WHERE name = 'uni_tbl_inv_item_release_doc_no'
                            AND object_id = OBJECT_ID('dbo.tbl_inv_item_release'))
             THEN 'yes' ELSE 'NO - INVESTIGATE' END;
GO
