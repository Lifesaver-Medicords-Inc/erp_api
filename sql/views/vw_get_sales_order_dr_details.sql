CREATE
OR ALTER VIEW [dbo].[vw_get_sales_order_dr_details] AS
SELECT a.order_details_id,
    a.based_id,
    a.item_id,
    a.item_code,
    a.item_description AS item_desc,
    a.qty,
    a.list_price AS unit_price,
    a.total_price AS total_cost,
    a.percent_discount AS discount,
    a.status,
    b.delivery_date,
    d.name as uom_name
FROM tbl_trans_sales_order_details a
    LEFT JOIN tbl_trans_sales_order b ON a.based_id = b.order_id
    LEFT JOIN tbl_setup_item c ON a.item_id = c.id
    LEFT JOIN tbl_setup_item_unit_measurement d ON c.unit_of_measure_id = d.id