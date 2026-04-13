CREATE
OR ALTER PROCEDURE [dbo].[sp_GetPurchaseOrderDetailsReceiving] @PurchaseId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT pod.id AS purchase_order_details_id,
    pod.item_id,
    pod.item_code,
    pod.item_description AS item_desc,
    pod.order_qty AS ordered_qty,
    pod.unit_of_measure AS ordered_uom,
    -- remaining order qty
    CASE
        WHEN (pod.order_qty - ISNULL(SUM(rrd.received_qty), 0)) < 0 THEN 0
        ELSE (pod.order_qty - ISNULL(SUM(rrd.received_qty), 0))
    END AS remaining_qty,
    pod.unit_of_measure AS remaining_uom
FROM tbl_purchasing_purchase_order_details pod
    LEFT JOIN tbl_inv_receiving_report_details rrd ON rrd.purchase_order_details_id = pod.id
    INNER JOIN tbl_purchasing_purchase_order po ON pod.based_id = po.id
WHERE pod.based_id = @PurchaseId
GROUP BY pod.id,
    pod.item_id,
    pod.item_code,
    pod.item_description,
    pod.order_qty,
    pod.unit_of_measure
HAVING (pod.order_qty - ISNULL(SUM(rrd.received_qty), 0)) > 0;
END TRY BEGIN CATCH THROW;
END CATCH
END;