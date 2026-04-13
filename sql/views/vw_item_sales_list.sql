ALTER VIEW [dbo].[vw_item_sales_list] AS
SELECT a.order_id as id,
    a.customer_id,
    b.item_id as based_id,
    a.date,
    a.doc as sales_order_no,
    c.name as customer_name
FROM tbl_trans_sales_order a
    LEFT JOIN tbl_trans_sales_order_details b ON a.order_id = b.based_id
    LEFT JOIN tbl_bpi c ON a.customer_id = c.id