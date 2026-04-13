ALTER VIEW [dbo].[vw_get_inventory_tracker] AS
SELECT a1.receiving_report_details_id AS id,
    a1.receiving_report_id AS rr_id,
    a1.purchase_order_details_id AS pod_id,
    a.item_code,
    c.name AS general_name,
    d.name AS brand,
    f1.long_description AS item_desc,
    f1.calibration,
    a1.bin_location AS location,
    a1.qty_in AS qty,
    a1.uom AS uom,
    f2.name AS warehouse_name,
    a1.warehouse_id,
    g.remarks,
    g.id AS rem_id,
    g.rrd_id
FROM tbl_inv_stocks_location a1
    LEFT JOIN tbl_inv_warehouse_receiving_report_details a ON a1.receiving_report_details_id = a.id
    LEFT JOIN tbl_setup_item b ON a.item_id = b.id
    LEFT JOIN tbl_setup_item_name c ON b.item_name_id = c.id
    LEFT JOIN tbl_setup_item_brand d ON b.item_brand_id = d.id
    LEFT JOIN tbl_inv_warehouse_receiving_report f ON a.receiving_report_id = f.id
    LEFT JOIN tbl_setup_item_additional_specs f1 ON b.id = f1.based_id
    LEFT JOIN tbl_inv_warehouse_name f2 ON a1.warehouse_id = f2.id
    LEFT JOIN tbl_inv_tracker g ON g.pod_id = a1.purchase_order_details_id
    AND g.rrd_id = a.id;