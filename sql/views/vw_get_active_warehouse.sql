IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_active_warehouse' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_active_warehouse] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_active_warehouse] AS
SELECT *
FROM tbl_inv_warehouse_name
WHERE is_inactive = 0;