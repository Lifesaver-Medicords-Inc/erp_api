ALTER VIEW [dbo].[vw_item_additional_specs] AS
SELECT a.*,
    b.name AS volume_unit_of_measure,
    c.name AS weight_unit_of_measure,
    f.pump_type_compatability_id,
    f.pump_type_compatability_names
FROM tbl_setup_item_additional_specs a
    LEFT JOIN tbl_setup_item_unit_measurement b ON a.volume_unit_of_measure_id = b.id
    LEFT JOIN tbl_setup_item_unit_measurement c ON a.weight_unit_of_measure_id = c.id
    LEFT JOIN (
        SELECT ISNULL(
                STUFF(
                    (
                        SELECT ',' + CAST(aa.pump_type_compatability_id AS NVARCHAR(MAX))
                        FROM tbl_setup_item_additional_specs_pump_type aa
                        WHERE aa.additional_specs_id = a.additional_specs_id FOR XML PATH('')
                    ),
                    1,
                    1,
                    ''
                ),
                ''
            ) AS pump_type_compatability_id,
            ISNULL(
                STUFF(
                    (
                        SELECT ',' + bb.code
                        FROM tbl_setup_item_additional_specs_pump_type aa
                            LEFT JOIN tbl_setup_item_pump_type bb ON aa.pump_type_compatability_id = bb.id
                        WHERE aa.additional_specs_id = a.additional_specs_id FOR XML PATH('')
                    ),
                    1,
                    1,
                    ''
                ),
                ''
            ) AS pump_type_compatability_names,
            a.additional_specs_id
        FROM tbl_setup_item_additional_specs_pump_type a
        GROUP BY a.additional_specs_id
    ) f ON a.id = f.additional_specs_id