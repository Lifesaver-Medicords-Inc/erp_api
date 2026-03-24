CREATE
OR ALTER VIEW [dbo].[GetProjectComponents] AS
SELECT tbl_trans_sales_quotation.id AS quotation_id,
    tbl_trans_sales_quotation.project_name,
    tbl_trans_sales_quotation.customer_id,
    tbl_trans_sales_project_item_set.item_set_id AS set_id,
    tbl_trans_sales_project_item_set.based_id AS item_set_based_on_quotation_id,
    tbl_trans_sales_project_item_set.tab_number AS item_set_name,
    a.bom_id,
    a.item_id,
    a.model,
    a.items_id,
    a.qty,
    a.based_id AS based_on_set_id,
    a.components,
    a.node_id,
    a.node_name,
    a.node_order,
    a.node_type,
    a.parent_node_id,
    vw_items.item_name,
    c.long_description AS short_desc,
    c.size,
    a.component_total,
    vw_items.unit_of_measure,
    GetBpiCustomer.customer_code AS customer_name,
    d.id as boq_id,
    d.remarks,
    d.notes,
    e.wiring_note
FROM tbl_trans_sales_quotation
    JOIN tbl_trans_sales_project_item_set ON tbl_trans_sales_quotation.id = tbl_trans_sales_project_item_set.based_id
    LEFT JOIN tbl_trans_sales_project_items a ON tbl_trans_sales_project_item_set.item_set_id = a.based_id
    LEFT JOIN vw_items ON a.item_id = vw_items.id
    LEFT JOIN tbl_setup_item_additional_specs c ON a.item_id = c.based_id
    LEFT JOIN GetBpiCustomer ON tbl_trans_sales_quotation.customer_id = GetBpiCustomer.bpi_id
    LEFT JOIN tbl_setup_item_boq_details d ON a.items_id = d.items_id
    LEFT JOIN tbl_wiring_user_inputs e ON a.items_id = e.items_id