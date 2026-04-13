ALTER PROCEDURE [dbo].[sp_GetPurchaseOrderDetailsInvoice] @PurchaseId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT pod.id AS purchase_order_details_id,
    pod.based_id,
    pod.item_code,
    pod.item_description,
    pod.unit_of_measure AS order_uom,
    -- remaining order qty
    CASE
        WHEN (pod.order_qty - ISNULL(SUM(ird.req_qty), 0)) < 0 THEN 0
        ELSE (pod.order_qty - ISNULL(SUM(ird.req_qty), 0))
    END AS order_qty,
    pod.unit_of_measure AS req_uom,
    pod.req_qty,
    pod.discounted_price AS unit_price,
    pod.total_price AS total_cost
FROM tbl_purchasing_purchase_order_details pod
    LEFT JOIN tbl_accounting_invoice_receipt_details ird ON ird.purchase_order_details_id = pod.id
WHERE pod.based_id = @PurchaseId
GROUP BY pod.id,
    pod.based_id,
    pod.item_code,
    pod.item_description,
    pod.unit_of_measure,
    pod.req_qty,
    pod.discounted_price,
    pod.total_price,
    pod.order_qty
HAVING (pod.order_qty - ISNULL(SUM(ird.req_qty), 0)) > 0;
END TRY BEGIN CATCH THROW;
END CATCH
END;