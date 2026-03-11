CREATE
OR ALTER PROCEDURE [dbo].[sp_GetSalesOrderInvoice] @CustomerId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT so.order_id AS sales_order_id,
    so.doc AS so_number,
    so.date AS doc_date,
    bpi.branch_name AS customer_name,
    so.total_amount_due AS net_amount,
    so.delivery_date AS date_deliver,
    so.sales_executive AS sales_person,
    so.total_amount_due AS total_sales
FROM dbo.tbl_trans_sales_order AS so
    INNER JOIN dbo.tbl_bpi_general AS bpi ON so.customer_id = bpi.id
WHERE so.customer_id = @CustomerId;
END TRY BEGIN CATCH THROW;
END CATCH
END;