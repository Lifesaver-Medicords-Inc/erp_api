ALTER VIEW [dbo].[vw_PumpSpecifications] AS
-- FLA and VOLTAGE are saved straight to tbl_setup_item_specs.fla_1/volt_1 (see
-- item_specs_service.go's CreateItemSpec/UpdateItemSpec) - never as title/value rows,
-- which is what this view used to query (tsis.title = 'FLA' OR tsis.title = 'VOLTAGE'
-- directly on tbl_setup_item_specs). That title/value shape is only ever used for the
-- separate "ADDITIONAL SPECS" fields (HORSEPOWER, SUCTION SIZE, etc., stored in the
-- child table tbl_setup_item_specs_template) - FLA/VOLTAGE never landed there, so the
-- old WHERE clause could never match a real row, for any item.
--
-- Kept the same output shape (item_id/item_name_id/item_name/template_name/item_title/
-- item_value) via UNION ALL so every existing consumer (models.ItemPumpSpecsView, and
-- Quotation.cs's FinalTxtBoxClicked which reads item_title = 'FLA'/'VOLTAGE') keeps
-- working unchanged.
SELECT tsi.id AS item_id,
    tsi.item_name_id,
    tin.name AS item_name,
    tsis.template AS template_name,
    'FLA' AS item_title,
    tsis.fla_1 AS item_value
FROM tbl_setup_item tsi
    LEFT JOIN tbl_setup_item_name tin ON tsi.item_name_id = tin.id
    INNER JOIN tbl_setup_item_specs tsis ON tsi.id = tsis.based_id
WHERE tsis.template = 'PUMP'
    AND tsis.fla_1 IS NOT NULL
    AND tsis.fla_1 <> ''

UNION ALL

SELECT tsi.id AS item_id,
    tsi.item_name_id,
    tin.name AS item_name,
    tsis.template AS template_name,
    'VOLTAGE' AS item_title,
    tsis.volt_1 AS item_value
FROM tbl_setup_item tsi
    LEFT JOIN tbl_setup_item_name tin ON tsi.item_name_id = tin.id
    INNER JOIN tbl_setup_item_specs tsis ON tsi.id = tsis.based_id
WHERE tsis.template = 'PUMP'
    AND tsis.volt_1 IS NOT NULL
    AND tsis.volt_1 <> '';
