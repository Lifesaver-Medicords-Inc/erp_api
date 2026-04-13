ALTER VIEW [dbo].[vw_item_production_list] AS
SELECT a.id as item_id,
    b.id as bom_id,
    c.item_id as bom_item_id,
    c.bom_qty,
    e.item_code,
    e.item_model
FROM tbl_setup_item a
    LEFT JOIN tbl_setup_item_bom b ON a.id = b.item_id
    LEFT JOIN tbl_setup_item_bom_details c ON b.id = c.item_bom_id
    LEFT JOIN tbl_setup_item e ON c.item_id = e.id