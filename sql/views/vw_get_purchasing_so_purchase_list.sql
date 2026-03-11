CREATE
OR ALTER VIEW [dbo].[vw_get_purchasing_so_purchase_list] AS
SELECT a.item_id,
    b.purchaser,
    STRING_AGG(CAST(a.based_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.order_details_id ASC
    ) AS order_ids,
    STRING_AGG(CAST(a.order_details_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.order_details_id ASC
    ) AS order_detail_ids,
    STRING_AGG(CAST(b.doc AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.order_details_id ASC
    ) AS sales_order_nos,
    STRING_AGG(CAST(b.project_name AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.order_details_id ASC
    ) AS project_names,
    STRING_AGG(CAST(b.sales_executive AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.order_details_id ASC
    ) AS sales_executives,
    -- Remaining qty per detail
    STRING_AGG(
        CAST(
            ISNULL(a.qty, 0) - ISNULL(a.allocated_qty, 0) AS NVARCHAR(MAX)
        ),
        ','
    ) WITHIN GROUP (
        ORDER BY a.order_details_id ASC
    ) AS qtys,
    STRING_AGG(CAST(a.list_price AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.order_details_id ASC
    ) AS unit_prices,
    STRING_AGG(CAST(a.percent_discount AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.order_details_id ASC
    ) AS discounts,
    MIN(a.item_code) AS item_code,
    MIN(a.item_description) AS item_description,
    MIN(d.name) AS unit_of_measure,
    MIN(e.name) AS item_name,
    MIN(f.name) AS item_brand,
    STRING_AGG(CONVERT(NVARCHAR, b.delivery_date, 23), ',') AS commitment_dates,
    -- Total remaining qty per item
    SUM(ISNULL(a.qty, 0) - ISNULL(a.allocated_qty, 0)) AS total_qty
FROM tbl_trans_sales_order_details a
    LEFT JOIN tbl_trans_sales_order b ON a.based_id = b.order_id
    LEFT JOIN tbl_setup_item c ON a.item_id = c.id
    LEFT JOIN tbl_setup_item_unit_measurement d ON c.unit_of_measure_id = d.id
    LEFT JOIN tbl_setup_item_name e ON c.item_name_id = e.id
    LEFT JOIN tbl_setup_item_brand f ON c.item_brand_id = f.id
WHERE a.status = 'CANVASS'
    AND b.status = 'ACTIVE'
    AND a.item_id <> 0
    AND ISNULL(a.qty, 0) - ISNULL(a.allocated_qty, 0) > 0 -- Filter for remaining qty > 0
GROUP BY a.item_id,
    b.purchaser;