CREATE
OR ALTER VIEW [dbo].[vw_bpi_items] AS
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
    b.short_desc,
    b.item_tangibility_type AS status_tangible,
    c.trade_value AS status_trade,
    a.branch_id AS bpi_item_branch_id,
    a.is_deleted AS item_is_deleted
FROM tbl_bpi_items a
    LEFT JOIN tbl_setup_item b ON a.item_id = b.id
    LEFT JOIN (
        SELECT ISNULL(STRING_AGG(b.value, ','), '') AS trade_value,
            b.based_id
        FROM tbl_item_trade_type a
            LEFT JOIN (
                SELECT bb.id,
                    aa.based_id,
                    aa.value
                FROM tbl_item_trade_type aa
                    LEFT JOIN tbl_setup_item bb ON aa.based_id = bb.id
                GROUP BY bb.id,
                    aa.based_id,
                    aa.value
            ) b ON a.based_id = b.based_id
            AND a.value = b.value
        GROUP by b.based_id
    ) c ON a.item_id = c.based_id