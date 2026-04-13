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