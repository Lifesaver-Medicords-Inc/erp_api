CREATE
OR ALTER PROCEDURE [dbo].[sp_GetSalesOrderDetailsInvoice] @SalesId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT sod.order_details_id AS sales_order_details_id,
    sod.based_id AS sales_order_id,
    sod.item_id,
    sod.item_code,
    sod.item_description,
    uom.name AS item_uom,
    -- remaining order qty
    CASE
        WHEN (sod.qty - ISNULL(SUM(sind.item_qty), 0)) < 0 THEN 0
        ELSE (sod.qty - ISNULL(SUM(sind.item_qty), 0))
    END AS item_qty,
    sod.percent_discount AS discount,
    sod.list_price AS unit_price,
    sod.total_price AS total_cost,
    so.delivery_date AS date_deliver
FROM tbl_trans_sales_order_details sod
    LEFT JOIN tbl_accounting_sales_invoice_details sind ON sind.sales_order_details_id = sod.order_details_id
    INNER JOIN tbl_trans_sales_order so ON sod.based_id = so.order_id
    INNER JOIN tbl_setup_item i ON sod.item_id = i.id
    INNER JOIN tbl_setup_item_unit_measurement uom ON i.unit_of_measure_id = uom.id
WHERE sod.based_id = @SalesId
GROUP BY sod.order_details_id,
    sod.based_id,
    sod.item_code,
    sod.item_description,
    uom.name,
    sod.qty,
    sod.item_id,
    sod.percent_discount,
    sod.list_price,
    sod.total_price,
    so.delivery_date
HAVING (sod.qty - ISNULL(SUM(sind.item_qty), 0)) > 0;
END TRY BEGIN CATCH THROW;
END CATCH
END;