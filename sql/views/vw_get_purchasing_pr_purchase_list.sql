CREATE
OR ALTER VIEW [dbo].[vw_get_purchasing_pr_purchase_list] AS
SELECT a.item_id,
    b.purchaser,
    STRING_AGG(CAST(a.pr_order_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.pr_order_id ASC
    ) AS purchase_requisition_detail_ids,
    STRING_AGG(CAST(a.based_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.pr_order_id ASC
    ) AS purchase_requisition_ids,
    STRING_AGG(CAST(b.doc_no AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.pr_order_id ASC
    ) AS purchase_requisition_nos,
    STRING_AGG(CAST(b.request_by AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.pr_order_id ASC
    ) AS requestors,
    STRING_AGG(CAST(b.department AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.pr_order_id ASC
    ) AS departments,
    -- Remaining qty per detail
    STRING_AGG(
        CAST(
            ISNULL(a.qty, 0) - ISNULL(a.allocated_qty, 0) AS NVARCHAR(MAX)
        ),
        ','
    ) WITHIN GROUP (
        ORDER BY a.pr_order_id ASC
    ) AS qtys,
    MIN(c.item_code) AS item_code,
    MIN(g.long_description) AS item_description,
    MIN(d.name) AS unit_of_measure,
    MIN(e.name) AS item_name,
    MIN(f.name) AS item_brand,
    STRING_AGG(CONVERT(NVARCHAR, b.date_required, 23), ',') AS commitment_dates,
    -- Total remaining qty per item
    SUM(ISNULL(a.qty, 0) - ISNULL(a.allocated_qty, 0)) AS total_qty
FROM tbl_purchasing_purchase_requisition_orders a
    LEFT JOIN tbl_purchasing_purchase_requisition b ON a.based_id = b.pr_id
    LEFT JOIN tbl_setup_item c ON a.item_id = c.id
    LEFT JOIN tbl_setup_item_unit_measurement d ON c.unit_of_measure_id = d.id
    LEFT JOIN tbl_setup_item_name e ON c.item_name_id = e.id
    LEFT JOIN tbl_setup_item_brand f ON c.item_brand_id = f.id
    LEFT JOIN tbl_setup_item_additional_specs g ON c.id = g.based_id
WHERE a.status = 'CANVASS'
    AND b.status = 'APPROVED'
    AND a.item_id <> 0
    AND ISNULL(a.qty, 0) - ISNULL(a.allocated_qty, 0) > 0 -- Filter for remaining qty > 0
GROUP BY a.item_id,
    b.purchaser;