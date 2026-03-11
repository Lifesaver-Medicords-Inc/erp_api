CREATE
OR ALTER VIEW [dbo].[vw_purchase_order_status] AS
SELECT po.id AS purchase_order_id,
    po.supplier_id,
    po.supplier_name,
    po.supplier_code,
    po.doc_no,
    pod.id AS pod_id,
    pod.item_id,
    pod.order_qty,
    ISNULL(SUM(CAST(wh.received_qty AS INT)), 0) AS total_received_qty,
    pod.order_qty - ISNULL(SUM(CAST(wh.received_qty AS INT)), 0) AS remaining_qty,
    pod.unit_of_measure AS ordered_uom,
    pod.item_code,
    ias.long_description AS item_description,
    CASE
        WHEN pod.order_qty - ISNULL(SUM(CAST(wh.received_qty AS INT)), 0) > 0 THEN 'Active'
        ELSE 'Closed'
    END AS po_status
FROM tbl_purchasing_purchase_order po
    INNER JOIN tbl_purchasing_purchase_order_details pod ON po.id = pod.based_id
    LEFT JOIN tbl_inv_warehouse_receiving_history wh ON pod.id = wh.purchase_order_id
    INNER JOIN tbl_setup_item i ON pod.item_id = i.id
    INNER JOIN tbl_setup_item_additional_specs ias ON i.id = ias.based_id
GROUP BY po.id,
    po.supplier_id,
    po.supplier_name,
    po.supplier_code,
    po.doc_no,
    pod.id,
    pod.item_id,
    pod.order_qty,
    pod.unit_of_measure,
    pod.item_code,
    ias.long_description;