CREATE
OR ALTER PROCEDURE [dbo].[GetModelList] AS BEGIN
SELECT a.*,
.name AS related_name,
    c.name AS related_brand
FROM tbl_setup_item_model a
    LEFT JOIN tbl_setup_item_name b ON a.item_name_id = b.id
    LEFT JOIN tbl_setup_item_brand c ON a.item_brand_id = c.id;
END;