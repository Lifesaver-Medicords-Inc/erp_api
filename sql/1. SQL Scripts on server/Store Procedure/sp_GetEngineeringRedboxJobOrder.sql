CREATE PROCEDURE [dbo].[sp_GetEngineeringRedboxJobOrder]
    @UserId INT
AS
BEGIN
    SET NOCOUNT ON;

    BEGIN TRY
        SELECT
            so.order_id AS id,
            bpi.branch_name AS client_name,
            so.doc AS document_no,
            COUNT(sod.order_details_id) AS items, -- count of rows in sod per order
            so.project_name AS project_name,
            so.delivery_date AS due_date,
            MAX(i.item_model) AS type -- pick one representative value if needed
        FROM dbo.tbl_trans_sales_order AS so
        INNER JOIN dbo.tbl_trans_sales_order_details AS sod
            ON so.order_id = sod.based_id
        INNER JOIN dbo.tbl_setup_item AS i
            ON sod.item_id = i.id
        INNER JOIN dbo.tbl_bpi_general AS bpi
            ON so.customer_id = bpi.id
        WHERE so.approved_by_id = @UserId
        GROUP BY 
            so.order_id,
            bpi.branch_name,
            so.doc,
            so.project_name,
            so.delivery_date;
    END TRY
    BEGIN CATCH
        THROW;
    END CATCH
END;
GO


