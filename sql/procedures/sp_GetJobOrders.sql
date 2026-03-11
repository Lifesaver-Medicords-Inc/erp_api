CREATE
OR ALTER PROCEDURE [dbo].[sp_GetJobOrders] @UserId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT ISNULL(jo.job_order_id, 0) AS id,
    sod.bom_id AS bom_id,
    sod.order_details_id AS order_details_id,
    so.date AS date,
    so.doc AS sales_order,
    jo.materials AS materials,
    sod.qty AS quantity,
    jo.due AS due,
    jo.a_engr AS a_engr,
    jo.item_rqst AS item_rqst,
    jo.status AS status,
    itn.name AS general_name,
    jo.serial_no AS serial_no,
    jo.report AS report,
    jo.report_base AS report_base,
    i.item_model AS type,
    ias.long_description AS item_desc
FROM dbo.tbl_trans_sales_order AS so
    INNER JOIN dbo.tbl_trans_sales_order_details AS sod ON so.order_id = sod.based_id
    INNER JOIN dbo.tbl_setup_item AS i ON sod.item_id = i.id
    INNER JOIN dbo.tbl_setup_item_name AS itn ON i.item_name_id = itn.id
    INNER JOIN tbl_setup_item_additional_specs ias ON ias.based_id = i.id
    LEFT JOIN dbo.tbl_trans_job_order AS jo ON sod.order_details_id = jo.order_details_id
WHERE so.approved_by_id = @UserId;
END TRY BEGIN CATCH THROW;
END CATCH
END;