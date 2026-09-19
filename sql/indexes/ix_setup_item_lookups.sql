-- vw_items and vw_item_bom_list join these two tables on columns that had no index, so
-- every read sorted or hashed the whole table first - and a sort/hash needs a SQL Server
-- memory grant. When the machine runs short of RAM, SQL Server gives memory back until
-- only a few MB of grant memory remain, and each of those reads queued for it (wait type
-- RESOURCE_SEMAPHORE, measured at 25 s to over 5 minutes). That was the sales quotation's
-- slow open: GET /setup/bom returns 8 rows and took 25 s. With these indexes the BOM view
-- is a handful of seeks that need no grant at all - 43 ms while the server was still
-- starved.
--
-- Idempotent and SQL Server 2012 compatible (no CREATE INDEX IF NOT EXISTS there).

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_setup_item_additional_specs_based_id'
                 AND object_id = OBJECT_ID('dbo.tbl_setup_item_additional_specs'))
    CREATE NONCLUSTERED INDEX IX_tbl_setup_item_additional_specs_based_id
        ON dbo.tbl_setup_item_additional_specs (based_id)
        INCLUDE (long_description);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_setup_item_trade_type_item_id'
                 AND object_id = OBJECT_ID('dbo.tbl_setup_item_trade_type'))
    CREATE NONCLUSTERED INDEX IX_tbl_setup_item_trade_type_item_id
        ON dbo.tbl_setup_item_trade_type (item_id)
        INCLUDE (trade_type_id);
GO
