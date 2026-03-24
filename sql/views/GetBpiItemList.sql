CREATE
OR ALTER VIEW [dbo].[GetBpiItemList] AS
SELECT a.id,
    c.long_description AS short_desc,
    a.item_code,
    b.name AS general_name,
    d.name AS item_brand_name
FROM tbl_setup_item a
    LEFT JOIN tbl_setup_item_name b ON a.item_name_id = b.id
    LEFT JOIN tbl_setup_item_additional_specs c ON a.id = c.based_id
    LEFT JOIN tbl_setup_item_brand d ON a.item_brand_id = d.id