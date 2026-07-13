CREATE OR ALTER PROCEDURE [dbo].[sp_GetSalesOrderInvoice] @CustomerId INT AS BEGIN
SET NOCOUNT ON;
SELECT so.order_id AS sales_order_id,
    so.doc AS so_number,
    dr.doc_no AS dr_number,
    so.date AS doc_date,
    bpi.branch_name AS customer_name,
    so.total_amount_due AS net_amount,
    so.delivery_date AS date_deliver,
    so.sales_executive AS sales_person,
    so.total_amount_due AS total_sales	
FROM dbo.tbl_trans_sales_order AS so
    INNER JOIN dbo.tbl_bpi_general AS bpi ON bpi.based_id = so.customer_id
    INNER JOIN dbo.tbl_dispatching_delivery_receipt AS dr ON dr.sales_order_id =  so.order_id
WHERE dr.customer_id = @CustomerId;
END