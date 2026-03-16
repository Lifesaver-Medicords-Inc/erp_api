CREATE VIEW [dbo].[vw_get_purchasing_pr_purchase_list] AS

SELECT
    a.item_id,
    b.purchaser,

    -- purchase_requisition_detail_ids
    STUFF((
        SELECT ',' + CAST(a2.pr_order_id AS NVARCHAR(MAX))
        FROM tbl_purchasing_purchase_requisition_orders a2
        WHERE a2.item_id = a.item_id
          AND a2.based_id = b.pr_id
        ORDER BY a2.pr_order_id
        FOR XML PATH(''), TYPE
    ).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS purchase_requisition_detail_ids,

    -- purchase_requisition_ids
    STUFF((
        SELECT ',' + CAST(a2.based_id AS NVARCHAR(MAX))
        FROM tbl_purchasing_purchase_requisition_orders a2
        WHERE a2.item_id = a.item_id
          AND a2.based_id = b.pr_id
        ORDER BY a2.pr_order_id
        FOR XML PATH(''), TYPE
    ).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS purchase_requisition_ids,

    -- purchase_requisition_nos
    STUFF((
        SELECT ',' + CAST(b2.doc_no AS NVARCHAR(MAX))
        FROM tbl_purchasing_purchase_requisition_orders a2
        INNER JOIN tbl_purchasing_purchase_requisition b2
            ON a2.based_id = b2.pr_id
        WHERE a2.item_id = a.item_id
          AND a2.based_id = b.pr_id
        ORDER BY a2.pr_order_id
        FOR XML PATH(''), TYPE
    ).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS purchase_requisition_nos,

    -- requestors
    STUFF((
        SELECT ',' + CAST(b2.request_by AS NVARCHAR(MAX))
        FROM tbl_purchasing_purchase_requisition_orders a2
        INNER JOIN tbl_purchasing_purchase_requisition b2
            ON a2.based_id = b2.pr_id
        WHERE a2.item_id = a.item_id
          AND a2.based_id = b.pr_id
        ORDER BY a2.pr_order_id
        FOR XML PATH(''), TYPE
    ).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS requestors,

    -- departments
    STUFF((
        SELECT ',' + CAST(b2.department AS NVARCHAR(MAX))
        FROM tbl_purchasing_purchase_requisition_orders a2
        INNER JOIN tbl_purchasing_purchase_requisition b2
            ON a2.based_id = b2.pr_id
        WHERE a2.item_id = a.item_id
          AND a2.based_id = b.pr_id
        ORDER BY a2.pr_order_id
        FOR XML PATH(''), TYPE
    ).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS departments,

    -- qtys (remaining qty per detail)
    STUFF((
        SELECT ',' + CAST(ISNULL(a2.qty,0) - ISNULL(a2.allocated_qty,0) AS NVARCHAR(MAX))
        FROM tbl_purchasing_purchase_requisition_orders a2
        WHERE a2.item_id = a.item_id
          AND a2.based_id = b.pr_id
        ORDER BY a2.pr_order_id
        FOR XML PATH(''), TYPE
    ).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS qtys,

    MIN(c.item_code) AS item_code,
    MIN(d.name) AS unit_of_measure,
    MIN(e.name) AS item_name,
    MIN(f.name) AS item_brand,

    -- commitment_dates
    STUFF((
        SELECT ',' + CONVERT(NVARCHAR(10), b2.date_required, 23)
        FROM tbl_purchasing_purchase_requisition_orders a2
        INNER JOIN tbl_purchasing_purchase_requisition b2
            ON a2.based_id = b2.pr_id
        WHERE a2.item_id = a.item_id
          AND a2.based_id = b.pr_id
        ORDER BY a2.pr_order_id
        FOR XML PATH(''), TYPE
    ).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS commitment_dates,

    -- total remaining qty per item
    SUM(ISNULL(a.qty,0) - ISNULL(a.allocated_qty,0)) AS total_qty

FROM tbl_purchasing_purchase_requisition_orders a
INNER JOIN tbl_purchasing_purchase_requisition b
    ON a.based_id = b.pr_id
LEFT JOIN tbl_setup_item c ON a.item_id = c.id
LEFT JOIN tbl_setup_item_unit_measurement d ON c.unit_of_measure_id = d.id
LEFT JOIN tbl_setup_item_name e ON c.item_name_id = e.id
LEFT JOIN tbl_setup_item_brand f ON c.item_brand_id = f.id

WHERE a.status = 'CANVASS'
  AND b.status = 'APPROVED'
  AND a.item_id <> 0
  AND ISNULL(a.qty,0) - ISNULL(a.allocated_qty,0) > 0

GROUP BY a.item_id, b.purchaser, b.pr_id;
GO