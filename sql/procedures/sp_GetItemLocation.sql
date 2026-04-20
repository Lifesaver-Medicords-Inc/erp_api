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