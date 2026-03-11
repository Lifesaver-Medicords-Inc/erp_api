CREATE
OR ALTER VIEW [dbo].[vw_get_purchasing_redbox_purchase_list] AS
SELECT *
FROM (
        SELECT a.order_id AS id,
            a.doc AS doc_no,
            a.project_name,
            a.delivery_date AS commitment_date,
            a.purchaser,
            STRING_AGG(CAST(b.order_details_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
                ORDER BY b.order_details_id ASC
            ) AS detail_ids,
            STRING_AGG(CAST(b.item_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
                ORDER BY b.order_details_id ASC
            ) AS item_ids,
            STRING_AGG(CAST(e.name AS NVARCHAR(MAX)), ',') WITHIN GROUP (
                ORDER BY b.order_details_id ASC
            ) AS item_names,
            MAX(c.name) AS customer,
            -- Ensures one customer name is selected per order
            'SO' AS order_type
        FROM tbl_trans_sales_order a
            LEFT JOIN tbl_trans_sales_order_details b ON a.order_id = b.based_id
            AND ISNULL(b.qty, 0) - ISNULL(b.allocated_qty, 0) > 0
            LEFT JOIN tbl_bpi c ON a.customer_id = c.id
            LEFT JOIN tbl_setup_item d ON b.item_id = d.id
            LEFT JOIN tbl_setup_item_name e ON d.item_name_id = e.id
        WHERE a.status = 'ACTIVE'
            AND COALESCE(b.status, '') = 'CANVASS'
        GROUP BY a.order_id,
            a.doc,
            a.delivery_date,
            a.project_name,
            a.purchaser
        UNION ALL
        SELECT a.pr_id AS id,
            a.doc_no,
            a.department AS project_name,
            a.date_required AS commitment_date,
            a.purchaser,
            STRING_AGG(CAST(b.pr_order_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
                ORDER BY b.pr_order_id ASC
            ) AS detail_ids,
            STRING_AGG(CAST(b.item_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
                ORDER BY b.pr_order_id ASC
            ) AS item_ids,
            STRING_AGG(CAST(d.name AS NVARCHAR(MAX)), ',') WITHIN GROUP (
                ORDER BY b.pr_order_id ASC
            ) AS item_names,
            a.request_by AS customer,
            'PR' AS order_type
        FROM tbl_purchasing_purchase_requisition a
            LEFT JOIN tbl_purchasing_purchase_requisition_orders b ON a.pr_id = b.based_id
            LEFT JOIN tbl_setup_item c ON b.item_id = c.id
            LEFT JOIN tbl_setup_item_name d ON c.item_name_id = d.id
        WHERE a.status = 'APPROVED'
            AND COALESCE(b.status, '') = 'CANVASS'
        GROUP BY a.pr_id,
            a.doc_no,
            a.department,
            a.date_required,
            a.purchaser,
            a.request_by
    ) a