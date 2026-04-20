ALTER VIEW [dbo].[vw_get_purchase_order_doc] AS
SELECT DISTINCT po.id AS purchase_order_id,
    po.doc_no AS po_doc_no
FROM tbl_purchasing_purchase_order po
    INNER JOIN tbl_purchasing_purchase_order_details pod ON po.id = pod.based_id
    LEFT JOIN (
        SELECT rrd.purchase_order_details_id,
            SUM(rrd.received_qty) AS total_received_qty
        FROM tbl_inv_receiving_report rr
            INNER JOIN tbl_inv_receiving_report_details rrd ON rr.id = rrd.receiving_report_id
        GROUP BY rrd.purchase_order_details_id
    ) agg_rrd ON pod.id = agg_rrd.purchase_order_details_id
WHERE (
        pod.order_qty - ISNULL(agg_rrd.total_received_qty, 0)
    ) > 0