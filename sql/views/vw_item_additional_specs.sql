CREATE
OR ALTER VIEW [dbo].[vw_item_additional_specs] AS
SELECT a.*,
    b.name AS volume_unit_of_measure,
    c.name AS weight_unit_of_measure,
    f.pump_type_compatability_id,
    f.pump_type_compatability_names
FROM tbl_setup_item_additional_specs a
    LEFT JOIN tbl_setup_item_unit_measurement b ON a.volume_unit_of_measure_id = b.id
    LEFT JOIN tbl_setup_item_unit_measurement c ON a.weight_unit_of_measure_id = c.id
    LEFT JOIN (
        SELECT ISNULL(STRING_AGG(b.pump_type_compatability_id, ','), '') AS pump_type_compatability_id,
            ISNULL(STRING_AGG(b.code, ','), '') AS pump_type_compatability_names,
            b.additional_specs_id
        FROM tbl_setup_item_additional_specs_pump_type a
            LEFT JOIN (
                SELECT bb.id,
                    aa.additional_specs_id,
                    bb.code,
                    aa.pump_type_compatability_id
                FROM tbl_setup_item_additional_specs_pump_type aa
                    LEFT JOIN tbl_setup_item_pump_type bb ON aa.pump_type_compatability_id = bb.id
                GROUP BY bb.id,
                    aa.additional_specs_id,
                    bb.code,
                    aa.pump_type_compatability_id
            ) b ON a.additional_specs_id = b.additional_specs_id
            AND a.pump_type_compatability_id = b.id
        GROUP by b.additional_specs_id
    ) f ON a.id = f.additional_specs_id