-- Phase 2 (quotation load): every child read behind the paged/list endpoints
-- is a based_id (or quotation_quick_id) seek. Without these, each of those
-- reads scans or hashes its whole table first - the same RESOURCE_SEMAPHORE
-- story as ix_setup_item_lookups.sql on a RAM-starved server.
--
-- Idempotent and SQL Server 2012 compatible (no CREATE INDEX IF NOT EXISTS there).
-- Applied automatically via migrations/run_sql.go's sql/indexes folder walk.
--
-- document_no needs nothing new: UQ_sales_quotation_doc_version already leads
-- with it (exact-match seeks for the versions endpoints).

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_trans_sales_quotation_quick_based_id'
                 AND object_id = OBJECT_ID('dbo.tbl_trans_sales_quotation_quick'))
    CREATE NONCLUSTERED INDEX IX_tbl_trans_sales_quotation_quick_based_id
        ON dbo.tbl_trans_sales_quotation_quick (based_id);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_trans_sales_quotation_selected_image_quick_id'
                 AND object_id = OBJECT_ID('dbo.tbl_trans_sales_quotation_selected_image'))
    CREATE NONCLUSTERED INDEX IX_tbl_trans_sales_quotation_selected_image_quick_id
        ON dbo.tbl_trans_sales_quotation_selected_image (quotation_quick_id);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_trans_sales_project_item_set_based_id'
                 AND object_id = OBJECT_ID('dbo.tbl_trans_sales_project_item_set'))
    CREATE NONCLUSTERED INDEX IX_tbl_trans_sales_project_item_set_based_id
        ON dbo.tbl_trans_sales_project_item_set (based_id);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_trans_sales_project_content_based_id'
                 AND object_id = OBJECT_ID('dbo.tbl_trans_sales_project_content'))
    CREATE NONCLUSTERED INDEX IX_tbl_trans_sales_project_content_based_id
        ON dbo.tbl_trans_sales_project_content (based_id);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_trans_sales_project_items_based_id'
                 AND object_id = OBJECT_ID('dbo.tbl_trans_sales_project_items'))
    CREATE NONCLUSTERED INDEX IX_tbl_trans_sales_project_items_based_id
        ON dbo.tbl_trans_sales_project_items (based_id);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_trans_sales_project_wiring_based_id'
                 AND object_id = OBJECT_ID('dbo.tbl_trans_sales_project_wiring'))
    CREATE NONCLUSTERED INDEX IX_tbl_trans_sales_project_wiring_based_id
        ON dbo.tbl_trans_sales_project_wiring (based_id);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_trans_sales_project_history_based_id'
                 AND object_id = OBJECT_ID('dbo.tbl_trans_sales_project_history'))
    CREATE NONCLUSTERED INDEX IX_tbl_trans_sales_project_history_based_id
        ON dbo.tbl_trans_sales_project_history (based_id);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_trans_sales_project_multiplier_based_id'
                 AND object_id = OBJECT_ID('dbo.tbl_trans_sales_project_multiplier'))
    CREATE NONCLUSTERED INDEX IX_tbl_trans_sales_project_multiplier_based_id
        ON dbo.tbl_trans_sales_project_multiplier (based_id);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes
               WHERE name = 'IX_tbl_trans_sales_project_content_adv_cfg_based_id'
                 AND object_id = OBJECT_ID('dbo.tbl_trans_sales_project_content_advanced_configuration'))
    CREATE NONCLUSTERED INDEX IX_tbl_trans_sales_project_content_adv_cfg_based_id
        ON dbo.tbl_trans_sales_project_content_advanced_configuration (based_id);
GO
