IF NOT EXISTS (SELECT 1 FROM sys.procedures WHERE name = 'sp_GetPurchaseOrderInvoice' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE PROCEDURE [dbo].[sp_GetPurchaseOrderInvoice] AS SET NOCOUNT ON;')
END
GO
ALTER PROCEDURE [dbo].[sp_GetPurchaseOrderInvoice] @SupplierId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT po.id,
    po.doc_no AS po_number,
    po.date AS doc_date,
    po.supplier_name AS supplier_name,
    po.total_amount_due AS total_amount_po
FROM dbo.tbl_purchasing_purchase_order AS po
WHERE po.supplier_id = @SupplierId;
END TRY BEGIN CATCH THROW;
END CATCH
END;