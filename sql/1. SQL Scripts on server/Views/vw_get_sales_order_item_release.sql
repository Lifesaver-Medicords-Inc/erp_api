CREATE VIEW [dbo].[vw_get_sales_order_item_release]
AS
SELECT 
    sod.order_details_id AS sales_order_details_id,
    sod.based_id AS sales_order_id,
	so.doc as ref_doc_no,
    sod.item_id,
    sod.item_description,
    sod.qty AS required_qty,
    i.unit_of_measure_id as required_uom_id,
	uom.name as required_uom,
    sod.delivery_preference
FROM tbl_trans_sales_order_details sod
LEFT JOIN tbl_trans_sales_order so 
	ON sod.based_id = so.order_id
LEFT JOIN tbl_setup_item i 
    ON sod.item_id = i.id
LEFT JOIN tbl_setup_item_unit_measurement uom
	ON i.unit_of_measure_id = uom.id


GO
