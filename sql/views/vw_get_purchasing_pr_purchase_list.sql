ALTER VIEW [dbo].[vw_get_purchasing_pr_purchase_list] AS
SELECT a.item_id,
    b.purchaser,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(a2.pr_order_id AS NVARCHAR(MAX))
                FROM tbl_purchasing_purchase_requisition_orders a2
                    LEFT JOIN tbl_purchasing_purchase_requisition b2 ON a2.based_id = b2.pr_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'APPROVED'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.pr_order_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS purchase_requisition_detail_ids,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(a2.based_id AS NVARCHAR(MAX))
                FROM tbl_purchasing_purchase_requisition_orders a2
                    LEFT JOIN tbl_purchasing_purchase_requisition b2 ON a2.based_id = b2.pr_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'APPROVED'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.pr_order_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS purchase_requisition_ids,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(b2.doc_no AS NVARCHAR(MAX))
                FROM tbl_purchasing_purchase_requisition_orders a2
                    LEFT JOIN tbl_purchasing_purchase_requisition b2 ON a2.based_id = b2.pr_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'APPROVED'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.pr_order_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS purchase_requisition_nos,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(b2.request_by AS NVARCHAR(MAX))
                FROM tbl_purchasing_purchase_requisition_orders a2
                    LEFT JOIN tbl_purchasing_purchase_requisition b2 ON a2.based_id = b2.pr_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'APPROVED'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.pr_order_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS requestors,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(b2.department AS NVARCHAR(MAX))
                FROM tbl_purchasing_purchase_requisition_orders a2
                    LEFT JOIN tbl_purchasing_purchase_requisition b2 ON a2.based_id = b2.pr_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'APPROVED'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.pr_order_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS departments,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(
                        ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) AS NVARCHAR(MAX)
                    )
                FROM tbl_purchasing_purchase_requisition_orders a2
                    LEFT JOIN tbl_purchasing_purchase_requisition b2 ON a2.based_id = b2.pr_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'APPROVED'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.pr_order_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS qtys,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CONVERT(NVARCHAR, b2.date_required, 23)
                FROM tbl_purchasing_purchase_requisition_orders a2
                    LEFT JOIN tbl_purchasing_purchase_requisition b2 ON a2.based_id = b2.pr_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'APPROVED'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0 FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS commitment_dates,
    MIN(c.item_code) AS item_code,
    MIN(g.long_description) AS item_description,
    MIN(d.name) AS unit_of_measure,
    MIN(e.name) AS item_name,
    MIN(f.name) AS item_brand,
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
    AND ISNULL(a.qty, 0) - ISNULL(a.allocated_qty, 0) > 0
GROUP BY a.item_id,
    b.purchaser;