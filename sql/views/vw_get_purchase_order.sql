CREATE
OR ALTER VIEW [dbo].[vw_get_purchase_order] AS WITH pod_with_received AS (
    SELECT pod.based_id AS purchase_order_id,
        pod.id AS pod_id,
        -- cast order quantity to INT
        CAST(pod.order_qty AS INT) AS order_qty,
        -- sum received quantity (cast to INT first)
        ISNULL(SUM(CAST(wh.received_qty AS INT)), 0) AS total_received_qty,
        -- compute balance (order_qty - received_qty)
        CAST(pod.order_qty AS INT) - ISNULL(SUM(CAST(wh.received_qty AS INT)), 0) AS balance_qty
    FROM tbl_purchasing_purchase_order_details pod
        LEFT JOIN tbl_inv_warehouse_receiving_history wh ON pod.id = wh.purchase_order_details_id
    GROUP BY pod.based_id,
        pod.id,
        pod.order_qty
),
po_status AS (
    SELECT purchase_order_id,
        SUM(balance_qty) AS total_balance_qty
    FROM pod_with_received
    GROUP BY purchase_order_id
)
SELECT po.id,
    po.supplier_id,
    po.supplier_name,
    po.supplier_code,
    po.doc_no
FROM tbl_purchasing_purchase_order po
    INNER JOIN po_status s ON po.id = s.purchase_order_id
WHERE s.total_balance_qty > 0;
-- Only POs that are NOT completed