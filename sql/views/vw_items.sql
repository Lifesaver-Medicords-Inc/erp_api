ALTER VIEW [dbo].[vw_items] AS
SELECT a.id,
    a.item_name_id,
    h.long_description,
    a.item_class_id,
    a.item_brand_id,
    a.unit_of_measure_id,
    a.item_tangibility_type,
    a.item_model,
    a.catalogue_year,
    a.price,
    a.item_code,
    a.is_stop_selling,
    b.name AS item_name,
    d.name AS item_class,
    e.name AS item_brand,
    f.name AS unit_of_measure,
    g.trade_type_id,
    g.trade_type_names
FROM tbl_setup_item a
    LEFT JOIN tbl_setup_item_name b ON a.item_name_id = b.id
    LEFT JOIN tbl_setup_item_class d ON a.item_class_id = d.id
    LEFT JOIN tbl_setup_item_brand e ON a.item_brand_id = e.id
    LEFT JOIN tbl_setup_item_unit_measurement f ON a.unit_of_measure_id = f.id
    LEFT JOIN tbl_setup_item_additional_specs h ON a.id = h.based_id
    LEFT JOIN (
        SELECT ISNULL(
                STUFF(
                    (
                        SELECT ',' + CAST(bb.id AS NVARCHAR(MAX))
                        FROM tbl_setup_item_trade_type aa
                            LEFT JOIN tbl_setup_item_type bb ON aa.trade_type_id = bb.id
                        WHERE aa.item_id = a.item_id FOR XML PATH('')
                    ),
                    1,
                    1,
                    ''
                ),
                ''
            ) AS trade_type_id,
            ISNULL(
                STUFF(
                    (
                        SELECT ',' + bb.name
                        FROM tbl_setup_item_trade_type aa
                            LEFT JOIN tbl_setup_item_type bb ON aa.trade_type_id = bb.id
                        WHERE aa.item_id = a.item_id FOR XML PATH('')
                    ),
                    1,
                    1,
                    ''
                ),
                ''
            ) AS trade_type_names,
            a.item_id
        FROM tbl_setup_item_trade_type a
        GROUP BY a.item_id
    ) g ON a.id = g.item_id