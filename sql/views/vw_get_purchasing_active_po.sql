IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_purchasing_active_po' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_purchasing_active_po] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_purchasing_active_po] AS
SELECT po.id AS id,
    po.doc_no AS doc_no,
    po.supplier_name AS supplier_name,
    CAST(po.total_amount_due AS VARCHAR) AS total_amount_due,
    -- get it directly from PO table
    FORMAT(DATEADD(DAY, 30, po.date), 'MM/dd/yyyy') AS lead_time
FROM tbl_purchasing_purchase_order po
WHERE po.total_amount_due > 0;
-- only active POs