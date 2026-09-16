-- Restored 2026-09-14. The Opportunities list (GET /api/sales/opportunity, models.OpportunityView)
-- reads this view, but its script was deleted in 72c1f55 (2026-04-13) and never re-added here, so
-- no database built since has it. Definition copied unchanged from Lightspeed_ERP_deploy_rehearsal.
IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_sales_opportunities' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_sales_opportunities] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_sales_opportunities] AS
SELECT
    ISNULL(b.id, 0) AS id,
    b.tag,
    a.version_no,
    a.is_finalized,
    a.document_no,
    a.customer_id,
    a.project_name,
    a.date,
    b.client_req,
    a.total_amount_due,
    b.last_update,
    b.stage,
    b.status,
    b.special_deal,
    a.gross_sales AS total,
    c.branch_name
FROM dbo.tbl_trans_sales_quotation a
LEFT JOIN dbo.tbl_trans_sales_opportunity b ON a.document_no = b.document_no AND a.version_no = b.version_no
LEFT JOIN dbo.tbl_bpi_general c ON a.customer_id = c.based_id
