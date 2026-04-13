ALTER VIEW [dbo].[vw_get_purchasing_so_purchase_list] AS
SELECT a.item_id,
    b.purchaser,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(a2.based_id AS NVARCHAR(MAX))
                FROM tbl_trans_sales_order_details a2
                    LEFT JOIN tbl_trans_sales_order b2 ON a2.based_id = b2.order_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'ACTIVE'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.order_details_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS order_ids,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(a2.order_details_id AS NVARCHAR(MAX))
                FROM tbl_trans_sales_order_details a2
                    LEFT JOIN tbl_trans_sales_order b2 ON a2.based_id = b2.order_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'ACTIVE'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.order_details_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS order_detail_ids,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(b2.doc AS NVARCHAR(MAX))
                FROM tbl_trans_sales_order_details a2
                    LEFT JOIN tbl_trans_sales_order b2 ON a2.based_id = b2.order_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'ACTIVE'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.order_details_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS sales_order_nos,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(b2.project_name AS NVARCHAR(MAX))
                FROM tbl_trans_sales_order_details a2
                    LEFT JOIN tbl_trans_sales_order b2 ON a2.based_id = b2.order_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'ACTIVE'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.order_details_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS project_names,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(b2.sales_executive AS NVARCHAR(MAX))
                FROM tbl_trans_sales_order_details a2
                    LEFT JOIN tbl_trans_sales_order b2 ON a2.based_id = b2.order_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'ACTIVE'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.order_details_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS sales_executives,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(
                        ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) AS NVARCHAR(MAX)
                    )
                FROM tbl_trans_sales_order_details a2
                    LEFT JOIN tbl_trans_sales_order b2 ON a2.based_id = b2.order_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'ACTIVE'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.order_details_id ASC FOR XML PATH('')
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
                SELECT ',' + CAST(a2.list_price AS NVARCHAR(MAX))
                FROM tbl_trans_sales_order_details a2
                    LEFT JOIN tbl_trans_sales_order b2 ON a2.based_id = b2.order_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'ACTIVE'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.order_details_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS unit_prices,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CAST(a2.percent_discount AS NVARCHAR(MAX))
                FROM tbl_trans_sales_order_details a2
                    LEFT JOIN tbl_trans_sales_order b2 ON a2.based_id = b2.order_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'ACTIVE'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0
                ORDER BY a2.order_details_id ASC FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS discounts,
    ISNULL(
        STUFF(
            (
                SELECT ',' + CONVERT(NVARCHAR, b2.delivery_date, 23)
                FROM tbl_trans_sales_order_details a2
                    LEFT JOIN tbl_trans_sales_order b2 ON a2.based_id = b2.order_id
                WHERE a2.item_id = a.item_id
                    AND b2.purchaser = b.purchaser
                    AND a2.status = 'CANVASS'
                    AND b2.status = 'ACTIVE'
                    AND a2.item_id <> 0
                    AND ISNULL(a2.qty, 0) - ISNULL(a2.allocated_qty, 0) > 0 FOR XML PATH('')
            ),
            1,
            1,
            ''
        ),
        ''
    ) AS commitment_dates,
    MIN(a.item_code) AS item_code,
    MIN(a.item_description) AS item_description,
    MIN(d.name) AS unit_of_measure,
    MIN(e.name) AS item_name,
    MIN(f.name) AS item_brand,
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
    AND ISNULL(a.qty, 0) - ISNULL(a.allocated_qty, 0) > 0
GROUP BY a.item_id,
    b.purchaser;