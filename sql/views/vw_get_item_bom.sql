IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_item_bom' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_item_bom] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_item_bom] AS
SELECT TOP 2147483647 si.id AS item_id,
    itas.long_description AS short_desc,
    si.item_code,
    si.item_model,
    itn.name AS general_name,
    itas.size,
    ium.name AS uom_name
FROM tbl_setup_item si
    LEFT JOIN tbl_setup_item_name itn ON si.item_name_id = itn.id
    LEFT JOIN tbl_setup_item_additional_specs itas ON si.id = itas.based_id
    LEFT JOIN tbl_setup_item_unit_measurement ium ON si.unit_of_measure_id = ium.id
WHERE si.item_code NOT IN (
        SELECT item_code
        FROM vw_get_bom_list
        WHERE item_code IS NOT NULL
    )
ORDER BY si.id ASC