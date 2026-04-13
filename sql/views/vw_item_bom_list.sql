ALTER VIEW [dbo].[vw_item_bom_list] AS
SELECT a.id,
    a.item_bom_id,
    a.item_id,
    b.item_model as size,
    a.unit_price,
    a.net_price,
    a.bom_qty,
    c.item_name,
    b.item_code,
    e.long_description AS short_desc,
    d.name as uom_name
FROM tbl_setup_item_bom_details a
    LEFT JOIN tbl_setup_item b ON a.item_id = b.id
    LEFT JOIN vw_items c ON a.item_id = c.id
    LEFT JOIN tbl_setup_item_unit_measurement d ON b.unit_of_measure_id = d.id
    LEFT JOIN tbl_setup_item_additional_specs e ON a.item_id = e.based_id