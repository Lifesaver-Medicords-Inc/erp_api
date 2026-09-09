IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'GetBpiItemList' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[GetBpiItemList] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[GetBpiItemList] AS
SELECT a.id,
    c.long_description AS short_desc,
    a.item_code,
    b.name AS general_name,
    d.name AS item_brand_name
FROM tbl_setup_item a
    LEFT JOIN tbl_setup_item_name b ON a.item_name_id = b.id
    LEFT JOIN tbl_setup_item_additional_specs c ON a.id = c.based_id
    LEFT JOIN tbl_setup_item_brand d ON a.item_brand_id = d.id