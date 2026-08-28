ALTER VIEW [dbo].[vw_bpi_item_list] AS
SELECT a.id,
    ISNULL(a.item_tangibility_type, '') AS item_type,
    ISNULL(a.item_tangibility_type, '') AS status_tangible,
    a.item_code,
    a.item_name AS general_name,
    a.item_model AS item_model_name,
    a.item_brand AS item_brand_name,
    b.long_description,
    b.long_description AS short_desc,
    ISNULL(c.trade_value, '') AS status_trade,
    a.price AS item_price
FROM vw_items a
    LEFT JOIN tbl_setup_item_additional_specs b on a.id = b.based_id
    LEFT JOIN (
        SELECT ISNULL(
                STUFF(
                    (
                        SELECT ',' + aa.value
                        FROM tbl_item_trade_type aa
                        WHERE aa.based_id = a.based_id FOR XML PATH('')
                    ),
                    1,
                    1,
                    ''
                ),
                ''
            ) AS trade_value,
            a.based_id
        FROM tbl_item_trade_type a
        GROUP BY a.based_id
    ) c ON a.id = c.based_id
