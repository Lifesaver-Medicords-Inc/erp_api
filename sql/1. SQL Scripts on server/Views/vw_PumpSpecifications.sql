CREATE VIEW [dbo].[vw_PumpSpecifications] AS
SELECT 
    tsi.id AS item_id,
    tsi.item_name_id,
    tin.name AS item_name, 
    tsis.template AS template_name,  
    tsis.title AS item_title,
    tsis.value AS item_value
FROM tbl_setup_item tsi
LEFT JOIN tbl_setup_item_name tin ON tsi.item_name_id = tin.id 
LEFT JOIN tbl_setup_item_specs tsis ON tsi.id = tsis.based_id
WHERE tsis.template = 'PUMP' 
  AND (tsis.title = 'FLA' OR tsis.title = 'VOLTAGE');


GO
