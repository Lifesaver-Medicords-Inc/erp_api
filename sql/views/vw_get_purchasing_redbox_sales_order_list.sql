CREATE
OR ALTER VIEW [dbo].[vw_get_purchasing_redbox_sales_order_list] AS
SELECT a.order_id as id,
    a.doc as doc_no,
    a.project_name,
    a.delivery_date as commitment_date,
    a.purchaser,
    STRING_AGG(CAST(b.order_details_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.order_id ASC
    ) AS detail_ids,
    STRING_AGG(CAST(b.item_id AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.order_id ASC
    ) AS item_ids,
    STRING_AGG(CAST(e.name AS NVARCHAR(MAX)), ',') WITHIN GROUP (
        ORDER BY a.order_id ASC
    ) AS item_names,
    MAX(c.name) as customer,
    'SO' AS order_type
FROM tbl_trans_sales_order a
    LEFT JOIN tbl_trans_sales_order_details b ON a.order_id = b.based_id
    LEFT JOIN tbl_bpi c ON a.customer_id = c.id
    LEFT JOIN tbl_setup_item d ON b.item_id = d.id
    LEFT JOIN tbl_setup_item_name e ON d.item_name_id = e.id
WHERE a.status = 'ACTIVE'
    AND b.status = 'CANVASS'
GROUP BY a.order_id,
    a.doc,
    a.delivery_date,
    a.project_name,
    a.purchaser