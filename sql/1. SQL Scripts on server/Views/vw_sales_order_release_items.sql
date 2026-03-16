CREATE VIEW [dbo].[vw_sales_order_release_items]
AS
SELECT 
    sod.order_details_id AS sales_order_details_id,
    sod.based_id AS sales_order_id,
    sod.item_id,
    sod.item_description,
    sod.qty AS required_qty,
    i.unit_of_measure_id as required_unit_of_measure_id,
    sod.delivery_preference
FROM tbl_trans_sales_order_details sod
LEFT JOIN tbl_setup_item i 
    ON sod.item_id = i.id


GO
