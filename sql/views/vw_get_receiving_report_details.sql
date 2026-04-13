CREATE
OR ALTER VIEW [dbo].[vw_get_receiving_report_details] AS
SELECT rrd.*,
    -- Total received qty per Purchase Order detail
    ISNULL(
        SUM(rrd.received_qty) OVER (PARTITION BY rrd.purchase_order_details_id),
        0
    ) AS total_received_qty,
    -- Remaining qty = original order qty - total received qty
    pod.order_qty - ISNULL(
        SUM(rrd.received_qty) OVER (PARTITION BY rrd.purchase_order_details_id),
        0
    ) AS remaining_qty,
    pod.unit_of_measure AS remaining_uom
FROM tbl_inv_receiving_report_details rrd
    INNER JOIN tbl_purchasing_purchase_order_details pod ON rrd.purchase_order_details_id = pod.id;
GO