ALTER VIEW [dbo].[vw_get_purchasing_redbox_purchase_requisition_list] AS
SELECT a.pr_id AS id,
    a.doc_no AS doc_no,
    a.department AS project_name,
    a.date_required AS commitment_date,
    a.purchaser,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(b2.pr_order_id AS NVARCHAR(MAX))
                FROM tbl_purchasing_purchase_requisition_orders b2
                WHERE b2.based_id = a.pr_id
                    AND b2.status = 'CANVASS'
                ORDER BY a.pr_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS detail_ids,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(b2.item_id AS NVARCHAR(MAX))
                FROM tbl_purchasing_purchase_requisition_orders b2
                WHERE b2.based_id = a.pr_id
                    AND b2.status = 'CANVASS'
                ORDER BY a.pr_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS item_ids,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(d2.name AS NVARCHAR(MAX))
                FROM tbl_purchasing_purchase_requisition_orders b2
                    LEFT JOIN tbl_setup_item c2 ON b2.item_id = c2.id
                    LEFT JOIN tbl_setup_item_name d2 ON c2.item_name_id = d2.id
                WHERE b2.based_id = a.pr_id
                    AND b2.status = 'CANVASS'
                ORDER BY a.pr_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS item_names,
    a.request_by AS customer,
    'PR' AS order_type
FROM tbl_purchasing_purchase_requisition a
    LEFT JOIN tbl_purchasing_purchase_requisition_orders b ON a.pr_id = b.based_id
    LEFT JOIN tbl_setup_item c ON b.item_id = c.id
    LEFT JOIN tbl_setup_item_name d ON c.item_name_id = d.id
WHERE a.status = 'APPROVED'
    AND b.status = 'CANVASS'
GROUP BY a.pr_id,
    a.request_by,
    a.doc_no,
    a.department,
    a.date_required,
    a.purchaser