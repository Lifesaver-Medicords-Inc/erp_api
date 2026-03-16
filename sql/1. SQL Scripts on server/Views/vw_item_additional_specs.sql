CREATE VIEW [dbo].[vw_item_additional_specs] AS

SELECT 
    a.*, 
    b.name AS volume_unit_of_measure,
    c.name AS weight_unit_of_measure,
    f.pump_type_compatability_id,
    f.pump_type_compatability_names

FROM tbl_setup_item_additional_specs a

LEFT JOIN tbl_setup_item_unit_measurement b 
    ON a.volume_unit_of_measure_id = b.id

LEFT JOIN tbl_setup_item_unit_measurement c 
    ON a.weight_unit_of_measure_id = c.id

LEFT JOIN (

    SELECT 
        t.additional_specs_id,

        -- pump_type_compatability_id
        STUFF((
            SELECT ',' + CAST(p.pump_type_compatability_id AS NVARCHAR(MAX))
            FROM tbl_setup_item_additional_specs_pump_type p
            WHERE p.additional_specs_id = t.additional_specs_id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,'') AS pump_type_compatability_id,

        -- pump_type_compatability_names
        STUFF((
            SELECT ',' + pt.code
            FROM tbl_setup_item_additional_specs_pump_type p
            LEFT JOIN tbl_setup_item_pump_type pt 
                ON p.pump_type_compatability_id = pt.id
            WHERE p.additional_specs_id = t.additional_specs_id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,'') AS pump_type_compatability_names

    FROM tbl_setup_item_additional_specs_pump_type t
    GROUP BY t.additional_specs_id

) f ON a.id = f.additional_specs_id


GO
