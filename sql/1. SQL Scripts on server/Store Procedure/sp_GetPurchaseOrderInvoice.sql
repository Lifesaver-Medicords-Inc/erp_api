CREATE PROCEDURE [dbo].[sp_GetPurchaseOrderInvoice]
    @SupplierId INT
AS
BEGIN
    SET NOCOUNT ON;

    BEGIN TRY
        SELECT
		    po.id,
			po.doc_no AS po_number,
			po.date AS doc_date,
			po.supplier_name AS supplier_name
        FROM dbo.tbl_purchasing_purchase_order AS po
        WHERE po.supplier_id = @SupplierId;
    END TRY
    BEGIN CATCH
        THROW;
    END CATCH
END;
GO


