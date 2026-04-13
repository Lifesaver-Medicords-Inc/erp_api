ALTER PROCEDURE [dbo].[sp_GetBinLocationItem] @ItemId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT a.bin_location AS location,
    a.warehouse_id,
    a.item_id,
    a.uom AS stock_uom,
    -- Compute new stock qty:
    -- a.qty_in - SUM(b.req_qty for same inventory stock id)
    a.qty_in - ISNULL(SUM(b.req_qty), 0) AS stock_qty
FROM dbo.tbl_inv_stocks_location AS a
    LEFT JOIN tbl_inv_stocks_location_history AS b ON a.id = b.inventory_stock_id
WHERE a.item_id = @ItemId
    AND a.bin_location IS NOT NULL
    AND LTRIM(RTRIM(a.bin_location)) <> ''
GROUP BY a.bin_location,
    a.warehouse_id,
    a.item_id,
    a.uom,
    a.qty_in;
END TRY BEGIN CATCH THROW;
END CATCH
END;