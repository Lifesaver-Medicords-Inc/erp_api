ALTER VIEW [dbo].[vw_get_active_sales_order] AS
SELECT a.order_id,
    a.date,
    a.delivery_date,
    a.status,
    c.name as customer_name
FROM tbl_trans_sales_order a --LEFT JOIN tbl_trans_sales_order_details b ON a.order_id = b.based_id
    LEFT JOIN tbl_bpi c ON a.customer_id = c.id
WHERE a.status = 'ACTIVE'