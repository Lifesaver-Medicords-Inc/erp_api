-- Phase 3 (quotation load): the per-item detail reads behind
-- GET /setup/item/detail/:id seek images by based_id. Applied automatically
-- via migrations/run_sql.go's sql/indexes folder walk.
--
-- Idempotent and SQL Server 2012 compatible (no CREATE INDEX IF NOT EXISTS there).

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_setup_item_image_based_id'
                 AND object_id = OBJECT_ID('dbo.tbl_setup_item_image'))
    CREATE NONCLUSTERED INDEX IX_tbl_setup_item_image_based_id
        ON dbo.tbl_setup_item_image (based_id);
GO
