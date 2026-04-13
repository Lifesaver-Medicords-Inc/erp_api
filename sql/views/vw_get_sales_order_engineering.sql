ALTER VIEW [dbo].[vw_get_sales_order_engineering] AS
SELECT so.order_id AS id,
    bpg.branch_name AS customer,
    bpi.tin AS tin,
    so.doc AS code,
    shipAddr.location AS delivery_to,
    billAddr.location AS bill_to,
    so.doc AS doc_no,
    so.date AS date,
    so.delivery_date AS delivery_date,
    so.ref_po AS reference_doc,
    so.status AS status
FROM dbo.tbl_trans_sales_order AS so
    LEFT JOIN dbo.tbl_bpi_general AS bpg ON so.customer_id = bpg.id
    LEFT JOIN dbo.tbl_bpi AS bpi ON bpg.based_id = bpi.id
    LEFT JOIN dbo.tbl_bpi_address AS shipAddr ON so.ship_to_id = shipAddr.id
    LEFT JOIN dbo.tbl_bpi_address AS billAddr ON so.bill_to_id = billAddr.id