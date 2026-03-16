CREATE VIEW [dbo].[vw_get_item_available_inventory] AS
SELECT
    a.warehouse_id,
    c.name AS warehouse_name,
    a.bin_location AS location,
    a.item_id,
    a.qty_in - ISNULL(SUM(b.req_qty), 0) AS stock_qty,
    a.uom AS stock_uom
FROM dbo.tbl_inv_stocks_location AS a
LEFT JOIN dbo.tbl_inv_stocks_location_history AS b
    ON a.id = b.inventory_stock_id
LEFT JOIN dbo.tbl_inv_warehouse_name AS c
    ON a.warehouse_id = c.id
GROUP BY
    a.bin_location,
    a.warehouse_id,
    a.item_id,
    a.uom,
    a.qty_in,
    c.name


GO
