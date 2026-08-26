-- §5.23 Production Report - the Warehouse Manager's acknowledgement queue. Unlike
-- sp_GetJobOrders (scoped to one engineer's approved SOs, for the Engineering app's
-- own 3-tab view), this is company-wide: every Job Order the engineer has marked
-- COMPLETE that the Warehouse Manager has not yet acknowledged. INNER JOIN to
-- tbl_trans_job_order (not LEFT, unlike sp_GetJobOrders) - a line with no job order
-- at all can never be "complete and awaiting WH ack".
CREATE OR ALTER PROCEDURE [dbo].[sp_GetPendingProductionReports] AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT jo.id AS id,
    sod.bom_id AS bom_id,
    so.order_id AS so_id,
    sod.order_details_id AS order_details_id,
    so.date AS date,
    so.doc AS sales_order,
    jo.materials AS materials,
    sod.qty AS quantity,
    jo.due AS due,
    jo.engr_id,
    jo.a_engr AS a_engr,
    jo.status AS status,
    sod.status AS so_item_status,
    itn.name AS general_name,
    jo.serial_no AS serial_no,
    jo.report AS report,
    jo.report_base AS report_base,
    i.item_model AS type,
    i.id AS item_id,
    ias.long_description AS item_desc
FROM dbo.tbl_trans_job_order AS jo
    INNER JOIN dbo.tbl_trans_sales_order_details AS sod ON jo.order_details_id = sod.order_details_id
    INNER JOIN dbo.tbl_trans_sales_order AS so ON sod.based_id = so.order_id
    INNER JOIN dbo.tbl_setup_item AS i ON sod.item_id = i.id
    INNER JOIN dbo.tbl_setup_item_name AS itn ON i.item_name_id = itn.id
    INNER JOIN dbo.tbl_setup_item_additional_specs ias ON ias.based_id = i.id
WHERE jo.status = 'COMPLETE'
    AND ISNULL(jo.is_wh_acknowledged, 0) = 0;
END TRY BEGIN CATCH THROW;
END CATCH
END;
