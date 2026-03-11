CREATE
OR ALTER PROCEDURE [dbo].[sp_GetSalesOrderDetailsFromJobOrders] @OrderId int AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT sod.item_code AS item_code,
    sod.item_description AS item_desc,
    sod.qty AS stock,
    sod.qty AS req_qty,
    sod.delivery_preference AS remark,
    sod.status AS status
FROM dbo.tbl_trans_sales_order_details AS sod
    INNER JOIN dbo.tbl_trans_sales_order AS so ON sod.based_id = so.order_id
WHERE sod.based_id = @OrderId;
END TRY BEGIN CATCH THROW;
END CATCH
END;