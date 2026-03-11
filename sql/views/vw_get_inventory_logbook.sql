CREATE
OR ALTER VIEW [dbo].[vw_get_inventory_logbook] AS
SELECT a1.id,
    CASE
        WHEN a1.receiving_report_id IS NULL
        AND a1.pick_activity_id IS NOT NULL THEN a4.sod_id
        ELSE a.pod_id
    END AS pod_id,
    a1.item_id,
    c.name AS general_name,
    d.name AS brand,
    ias.long_description AS item_description,
    b.item_code,
    b.item_model,
    g.name AS item_category,
    g2.calibration,
    a1.bin_location AS location,
    /* qty_in = SUM of all qty_in for same item_id, bin_location, warehouse_id */
    (
        SELECT ISNULL(SUM(a2.qty_in), 0)
        FROM tbl_inv_stocks_location a2
        WHERE a2.item_id = a1.item_id
            AND a2.bin_location = a1.bin_location
            AND a2.warehouse_id = a1.warehouse_id
    ) AS qty_in,
    /* qty_out = SUM of all req_qty in ish for this stock */
    (
        SELECT ISNULL(SUM(ish2.req_qty), 0)
        FROM tbl_inv_stocks_location_history ish2
        WHERE ish2.inventory_stock_id = a1.id
    ) AS qty_out,
    a1.uom,
    -- rr_no logic
    CASE
        WHEN a1.receiving_report_id IS NULL
        AND a1.pick_activity_id IS NOT NULL THEN a3.doc_no
        ELSE a2.doc
    END AS rr_no,
    -- po_no logic
    CASE
        WHEN a1.receiving_report_id IS NULL
        AND a1.pick_activity_id IS NOT NULL THEN a3.reference_so
        ELSE a2.ref_doc
    END AS po_no,
    a1.supplier_name,
    a1.date_received AS date,
    h.remarks
FROM tbl_inv_stocks_location a1
    LEFT JOIN tbl_inv_warehouse_receiving_report_details2 a ON a1.receiving_report_details_id = a.id
    LEFT JOIN tbl_inv_warehouse_receiving_report2 a2 ON a1.receiving_report_id = a2.id
    LEFT JOIN tbl_inv_pick_activity a3 ON a1.pick_activity_id = a3.id
    LEFT JOIN tbl_inv_pick_activity_details a4 ON a1.pick_activity_details_id = a4.id
    LEFT JOIN tbl_setup_item b ON a.item_id = b.id
    LEFT JOIN tbl_setup_item_name c ON b.item_name_id = c.id
    LEFT JOIN tbl_setup_item_brand d ON b.item_brand_id = d.id
    LEFT JOIN tbl_setup_item_class g ON b.item_class_id = g.id
    LEFT JOIN tbl_setup_item_additional_specs g2 ON b.id = g2.based_id
    LEFT JOIN tbl_setup_item_additional_specs ias ON b.id = ias.based_id
    LEFT JOIN tbl_inv_tracker h ON h.pod_id = a1.purchase_order_details_id
    AND h.rrd_id = a.id;