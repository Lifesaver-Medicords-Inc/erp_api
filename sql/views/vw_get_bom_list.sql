CREATE
OR ALTER VIEW [dbo].[vw_get_bom_list] AS
SELECT a.id,
    a.item_id,
    a.production_qty,
    a.production_type,
    b.short_desc,
    a.man_days,
    a.labor_rate,
    a.production_cost,
    b.item_model,
    b.item_code,
    c.name as general_name
FROM tbl_setup_item_bom a
    LEFT JOIN tbl_setup_item b ON a.item_id = b.id
    LEFT JOIN tbl_setup_item_name c ON b.item_name_id = c.id