CREATE
OR ALTER VIEW [dbo].[vw_get_purchasing_redbox_purchase_requisition_list] AS
SELECT a.pr_id as id,
    a.doc_no as doc_no,
    a.department as project_name,
    a.date_required as commitment_date,
    a.purchaser,
    STRING_AGG(CAST(b.pr_order_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.pr_id ASC
    ) AS detail_ids,
    STRING_AGG(CAST(b.item_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.pr_id ASC
    ) AS item_ids,
    STRING_AGG(CAST (d.name AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.pr_id ASC
    ) AS item_names,
    a.request_by as customer,
    'PR' AS order_type
FROM tbl_purchasing_purchase_requisition a
    LEFT JOIN tbl_purchasing_purchase_requisition_orders b ON a.pr_id = b.based_id
    LEFT JOIN tbl_setup_item c ON b.item_id = c.id
    LEFT JOIN tbl_setup_item_name d ON c.item_name_id = d.id
WHERE a.status = 'APPROVED'
    and b.status = 'CANVASS'
GROUP BY a.pr_id,
    a.request_by,
    a.doc_no,
    a.department,
    a.date_required,
    a.purchaser