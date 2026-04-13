ALTER VIEW [dbo].[vw_bpi_items] AS
SELECT a.id AS bpi_item_id,
    a.based_id AS bpi_item_based_id,
    a.payment_terms_id,
    a.item_account_id,
    a.tax_code,
    a.item_tax_code,
    a.price,
    a.notes,
    a.item_id,
    b.item_code,
    d.long_description AS short_desc,
    b.item_tangibility_type AS status_tangible,
    c.trade_value AS status_trade,
    a.branch_id AS bpi_item_branch_id,
    a.is_deleted AS item_is_deleted
FROM tbl_bpi_items a
    LEFT JOIN tbl_setup_item b ON a.item_id = b.id
    LEFT JOIN tbl_setup_item_additional_specs d ON a.id = d.based_id
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
    ) c ON a.item_id = c.based_id