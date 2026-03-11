CREATE
OR ALTER VIEW [dbo].[vw_get_sales_order_ir] AS WITH sod_with_received AS (
    SELECT sod.based_id AS sales_order_id,
        sod.order_details_id AS sod_id,
        sod.item_id,
        sod.item_description,
        CAST(sod.qty AS INT) AS order_qty,
        ISNULL(SUM(CAST(ih.req_qty AS INT)), 0) AS total_required_qty,
        CAST(sod.qty AS INT) - ISNULL(SUM(CAST(ih.req_qty AS INT)), 0) AS balance_qty
    FROM tbl_trans_sales_order_details sod
        LEFT JOIN tbl_inv_item_request_history ih ON sod.order_details_id = ih.sod_id
    GROUP BY sod.based_id,
        sod.order_details_id,
        sod.qty,
        sod.item_id,
        sod.item_description -- FIXED
),
so_status AS (
    SELECT sales_order_id,
        sod_id,
        item_id,
        item_description,
        SUM(total_required_qty) AS total_required_qty,
        SUM(balance_qty) AS total_balance_qty
    FROM sod_with_received
    GROUP BY sales_order_id,
        sod_id,
        item_id,
        item_description -- FIXED
)
SELECT so.order_id AS so_id,
    s.sod_id,
    s.item_id,
    so.doc AS ref_doc,
    s.item_description,
    s.total_required_qty,
    s.total_balance_qty AS order_qty,
    s.total_balance_qty AS req_qty,
    um.name AS req_uom
FROM tbl_trans_sales_order so
    INNER JOIN so_status s ON so.order_id = s.sales_order_id
    INNER JOIN tbl_setup_item i ON s.item_id = i.id
    INNER JOIN tbl_setup_item_unit_measurement um ON i.unit_of_measure_id = um.id
WHERE s.total_balance_qty > 0;