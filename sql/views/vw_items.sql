CREATE
OR ALTER VIEW [dbo].[vw_items] AS
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
        SELECT ISNULL(STRING_AGG(b.trade_type_id, ','), '') AS trade_type_id,
            ISNULL(STRING_AGG(b.name, ','), '') AS trade_type_names,
            b.item_id
        FROM tbl_setup_item_trade_type a
            LEFT JOIN (
                SELECT aa.item_id,
                    bb.id AS trade_type_id,
                    bb.name
                FROM tbl_setup_item_trade_type aa
                    LEFT JOIN tbl_setup_item_type bb ON aa.trade_type_id = bb.id
                GROUP BY aa.item_id,
                    bb.id,
                    bb.name
            ) b ON a.item_id = b.item_id
            AND a.trade_type_id = b.trade_type_id
        GROUP BY b.item_id
    ) g ON a.id = g.item_id