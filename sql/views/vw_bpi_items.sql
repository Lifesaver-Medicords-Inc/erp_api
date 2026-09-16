IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_bpi_items' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_bpi_items] AS SELECT 1 AS placeholder')
END
GO
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
    -- TYPE and DESCRIPTION on the BPI Items tab. The grid binds to these two names, and the
    -- view never returned them, so every saved item came back with both columns blank.
    ISNULL(b.item_tangibility_type, '') AS item_type,
    d.long_description,
    d.long_description AS short_desc,
    b.item_tangibility_type AS status_tangible,
    c.trade_value AS status_trade,
    a.branch_id AS bpi_item_branch_id,
    a.is_deleted AS item_is_deleted
FROM tbl_bpi_items a
    LEFT JOIN tbl_setup_item b ON a.item_id = b.id
    -- The item's description. Joined on the ITEM (a.item_id); it used to join on a.id, the
    -- BPI item row's own id, which matched some unrelated item's specs or nothing at all.
    -- TOP 1 so an item carrying more than one specs row cannot repeat the supplier's item.
    OUTER APPLY (
        SELECT TOP 1 s.long_description
        FROM tbl_setup_item_additional_specs s
        WHERE s.based_id = a.item_id
        ORDER BY s.id
    ) d
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