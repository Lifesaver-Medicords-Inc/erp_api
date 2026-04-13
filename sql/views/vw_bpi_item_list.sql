ALTER VIEW [dbo].[vw_bpi_item_list] AS
SELECT a.id,
    ISNULL(a.item_tangibility_type, '') AS item_type,
    a.item_code,
    a.item_name AS general_name,
    a.item_model AS item_model_name,
    a.item_brand AS item_brand_name,
    b.long_description,
    a.price AS item_price
FROM vw_items a
    LEFT JOIN tbl_setup_item_additional_specs b on a.id = b.based_id