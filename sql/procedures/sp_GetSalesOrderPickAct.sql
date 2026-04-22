ALTER PROCEDURE [dbo].[sp_GetSalesOrderPickAct] @SalesId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT so.order_id AS sales_order_id,
    bpi.branch_name AS customer,
    bpi.customer_code,
    so.sales_executive AS sales_person
FROM dbo.tbl_trans_sales_order AS so
    INNER JOIN tbl_bpi_general AS bpi ON so.customer_id = bpi.id
WHERE so.order_id = @SalesId;
END TRY BEGIN CATCH THROW;
END CATCH
END;