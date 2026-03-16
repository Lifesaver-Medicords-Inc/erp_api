CREATE VIEW [dbo].[vw_get_purchasing_redbox_purchase_list] AS
SELECT * FROM (

SELECT 
    a.order_id AS id,
    a.doc AS doc_no,
    a.project_name,
    a.delivery_date AS commitment_date,
    a.purchaser,

    -- detail_ids
    STUFF((
        SELECT ',' + CAST(b2.order_details_id AS NVARCHAR(MAX))
        FROM tbl_trans_sales_order_details b2
        WHERE b2.based_id = a.order_id
        AND ISNULL(b2.qty,0) - ISNULL(b2.allocated_qty,0) > 0
        AND COALESCE(b2.status,'') = 'CANVASS'
        ORDER BY b2.order_details_id
        FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'),1,1,'') AS detail_ids,

    -- item_ids
    STUFF((
        SELECT ',' + CAST(b2.item_id AS NVARCHAR(MAX))
        FROM tbl_trans_sales_order_details b2
        WHERE b2.based_id = a.order_id
        AND ISNULL(b2.qty,0) - ISNULL(b2.allocated_qty,0) > 0
        AND COALESCE(b2.status,'') = 'CANVASS'
        ORDER BY b2.order_details_id
        FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'),1,1,'') AS item_ids,

    -- item_names
    STUFF((
        SELECT ',' + CAST(e2.name AS NVARCHAR(MAX))
        FROM tbl_trans_sales_order_details b2
        LEFT JOIN tbl_setup_item d2 ON b2.item_id = d2.id
        LEFT JOIN tbl_setup_item_name e2 ON d2.item_name_id = e2.id
        WHERE b2.based_id = a.order_id
        AND ISNULL(b2.qty,0) - ISNULL(b2.allocated_qty,0) > 0
        AND COALESCE(b2.status,'') = 'CANVASS'
        ORDER BY b2.order_details_id
        FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'),1,1,'') AS item_names,

    c.name AS customer,
    'SO' AS order_type

FROM tbl_trans_sales_order a
LEFT JOIN tbl_bpi c ON a.customer_id = c.id
WHERE a.status = 'ACTIVE'


UNION ALL


SELECT 
    a.pr_id AS id,
    a.doc_no,
    a.department AS project_name,
    a.date_required AS commitment_date,
    a.purchaser,

    -- detail_ids
    STUFF((
        SELECT ',' + CAST(b2.pr_order_id AS NVARCHAR(MAX))
        FROM tbl_purchasing_purchase_requisition_orders b2
        WHERE b2.based_id = a.pr_id
        AND COALESCE(b2.status,'') = 'CANVASS'
        ORDER BY b2.pr_order_id
        FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'),1,1,'') AS detail_ids,

    -- item_ids
    STUFF((
        SELECT ',' + CAST(b2.item_id AS NVARCHAR(MAX))
        FROM tbl_purchasing_purchase_requisition_orders b2
        WHERE b2.based_id = a.pr_id
        AND COALESCE(b2.status,'') = 'CANVASS'
        ORDER BY b2.pr_order_id
        FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'),1,1,'') AS item_ids,

    -- item_names
    STUFF((
        SELECT ',' + CAST(d2.name AS NVARCHAR(MAX))
        FROM tbl_purchasing_purchase_requisition_orders b2
        LEFT JOIN tbl_setup_item c2 ON b2.item_id = c2.id
        LEFT JOIN tbl_setup_item_name d2 ON c2.item_name_id = d2.id
        WHERE b2.based_id = a.pr_id
        AND COALESCE(b2.status,'') = 'CANVASS'
        ORDER BY b2.pr_order_id
        FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'),1,1,'') AS item_names,

    a.request_by AS customer,
    'PR' AS order_type

FROM tbl_purchasing_purchase_requisition a
WHERE a.status = 'APPROVED'

) a


GO
