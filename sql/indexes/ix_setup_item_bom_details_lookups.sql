-- The BOM screen reads a page of 20 BOMs and then the component lines of those 20
-- (GET /setup/bom/paged, 2026-09-26). tbl_setup_item_bom_details had only its primary key, so
-- finding a BOM's lines meant reading all of them - 130,380 on test_fresh. This makes it a seek.
--
-- Idempotent and SQL Server 2012 compatible (no CREATE INDEX IF NOT EXISTS there).

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_setup_item_bom_details_item_bom_id'
                 AND object_id = OBJECT_ID('dbo.tbl_setup_item_bom_details'))
    CREATE NONCLUSTERED INDEX IX_tbl_setup_item_bom_details_item_bom_id
        ON dbo.tbl_setup_item_bom_details (item_bom_id);
GO
