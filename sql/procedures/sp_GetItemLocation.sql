IF NOT EXISTS (SELECT 1 FROM sys.procedures WHERE name = 'sp_GetItemLocation' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE PROCEDURE [dbo].[sp_GetItemLocation] AS SET NOCOUNT ON;')
END
GO
ALTER PROCEDURE [dbo].[sp_GetItemLocation] @ItemId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT its.id AS bin_id,
    its.bin_location,
    its.warehouse_id,
    its.item_id,
    its.stock_qty,
    its.stock_uom
FROM dbo.tbl_inv_item_stocks AS its
WHERE its.item_id = @ItemId
    AND its.is_active = 1;
END TRY BEGIN CATCH THROW;
END CATCH
END;