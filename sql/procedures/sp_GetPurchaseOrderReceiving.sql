IF NOT EXISTS (SELECT 1 FROM sys.procedures WHERE name = 'sp_GetPurchaseOrderReceiving' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE PROCEDURE [dbo].[sp_GetPurchaseOrderReceiving] AS SET NOCOUNT ON;')
END
GO
ALTER PROCEDURE [dbo].[sp_GetPurchaseOrderReceiving] @PurchaseId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT po.id AS purchase_order_id,
    po.supplier_id,
    po.supplier_name AS supplier,
    po.supplier_code
FROM dbo.tbl_purchasing_purchase_order AS po
WHERE po.id = @PurchaseId;
END TRY BEGIN CATCH THROW;
END CATCH
END;