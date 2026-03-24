CREATE
OR ALTER VIEW [dbo].[vw_get_all_item_bom] AS
SELECT a.id AS item_id,
    c.long_description AS short_desc,
    a.item_code,
    a.item_model,
    b.name AS general_name,
    c.size,
    d.name AS uom_name
FROM tbl_setup_item a
    LEFT JOIN tbl_setup_item_name b ON a.item_name_id = b.id
    LEFT JOIN tbl_setup_item_additional_specs c ON a.id = c.based_id
    LEFT JOIN tbl_setup_item_unit_measurement d ON a.unit_of_measure_id = d.id