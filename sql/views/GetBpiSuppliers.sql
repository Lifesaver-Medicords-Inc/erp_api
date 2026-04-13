ALTER VIEW [dbo].[GetBpiSuppliers] AS
SELECT tbl_bpi_general.id,
    tbl_bpi_general.supplier_code,
    tbl_bpi_items.based_id,
    tbl_bpi_items.item_id,
    vw_items.item_name,
    tbl_bpi_items.price
FROM tbl_bpi_items
    JOIN tbl_bpi_general ON tbl_bpi_items.based_id = tbl_bpi_general.based_id
    LEFT JOIN vw_items ON tbl_bpi_items.item_id = vw_items.id;