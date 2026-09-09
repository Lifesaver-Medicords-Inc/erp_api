IF NOT EXISTS (SELECT 1 FROM sys.procedures WHERE name = 'sp_GetWarehouseAreaReceiving' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE PROCEDURE [dbo].[sp_GetWarehouseAreaReceiving] AS SET NOCOUNT ON;')
END
GO
ALTER PROCEDURE [dbo].[sp_GetWarehouseAreaReceiving] @WarehouseId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT wa.id AS warehouse_area_id,
    wa.zone,
    wa.area,
    wa.rack,
    wa.level,
    wa.bins
FROM dbo.tbl_inv_warehouse_area AS wa
WHERE wa.warehouse_name_id = @WarehouseId;
END TRY BEGIN CATCH THROW;
END CATCH
END;