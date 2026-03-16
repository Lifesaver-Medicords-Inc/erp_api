CREATE VIEW [dbo].[vw_get_purchasing_closed_po] AS


SELECT
    po.id AS id,
    po.doc_no AS doc_no,
    po.supplier_name AS supplier_name,
    CAST(po.total_amount_due AS VARCHAR) AS total_amount_due,  -- total from PO table
    FORMAT(DATEADD(DAY, 30, po.date), 'MM/dd/yyyy') AS lead_time,
    rr.id AS receiving_report_id,
    rr.doc AS receiving_report_no
FROM tbl_purchasing_purchase_order po
INNER JOIN tbl_purchasing_purchase_order_details pod
    ON po.id = pod.based_id
LEFT JOIN (
    SELECT purchase_order_details_id, SUM(CAST(received_qty AS INT)) AS received_qty
    FROM tbl_inv_warehouse_receiving_history
    GROUP BY purchase_order_details_id
) wh
    ON pod.id = wh.purchase_order_details_id
LEFT JOIN tbl_inv_warehouse_receiving_report rr
    ON rr.purchase_order_id = po.id
GROUP BY
    po.id,
    po.doc_no,
    po.supplier_name,
    po.date,
    po.total_amount_due,
    rr.id,
    rr.doc
HAVING SUM(pod.order_qty - ISNULL(wh.received_qty, 0)) = 0;  -- only fully received items


GO
