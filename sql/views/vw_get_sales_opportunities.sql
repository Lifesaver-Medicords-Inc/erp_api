-- Restored 2026-09-14. The Opportunities list (GET /api/sales/opportunity, models.OpportunityView)
-- reads this view, but its script was deleted in 72c1f55 (2026-04-13) and never re-added here, so
-- no database built since has it. Definition copied unchanged from Lightspeed_ERP_deploy_rehearsal.
IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_sales_opportunities' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_sales_opportunities] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_sales_opportunities] AS
WITH latest AS (
    -- One row per quote: its latest copy - highest version, then highest sub-version, then
    -- the FQ over the draft it was finalized from (the FQ is inserted after, so its id is
    -- higher). Keyed on the number without its Q#/FQ# prefix so a quote and its FQ are one
    -- quote. Every Edit session is its own row since versioning (2026-09-25); without this
    -- each one would show again here and be counted again.
    SELECT q.*,
        ROW_NUMBER() OVER (
            PARTITION BY REPLACE(REPLACE(q.document_no, 'FQ#', ''), 'Q#', '')
            ORDER BY TRY_CAST(q.version_no AS int) DESC,
                     TRY_CAST(q.sub_version_no AS int) DESC,
                     q.id DESC
        ) AS rn
    FROM tbl_trans_sales_quotation q
)
SELECT
    ISNULL(b.id, 0) AS id,
    b.tag,
    a.version_no,
    a.sub_version_no,
    a.document_no + '.' + ISNULL(NULLIF(a.version_no, ''), '1') + '.' + ISNULL(NULLIF(a.sub_version_no, ''), '0') AS quote_code,
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
FROM latest a
LEFT JOIN dbo.tbl_trans_sales_opportunity b ON a.document_no = b.document_no AND a.version_no = b.version_no
-- ONE company name per quote. customer_id is the business partner, and a partner has a
-- tbl_bpi_general row per branch, so a plain join listed the quote once per branch - a
-- partner with a customer branch and a supplier branch showed every quote twice
-- (user-reported 2026-09-25). The customer branch, since a quote goes to a customer; then
-- the main branch; then the oldest - which is also the row the Quotation screen shows.
OUTER APPLY (
    SELECT TOP 1 g.branch_name
      FROM dbo.tbl_bpi_general g
     WHERE g.based_id = a.customer_id
     ORDER BY CASE WHEN ISNULL(g.customer_code, '') <> '' THEN 0 ELSE 1 END,
              CASE WHEN g.is_main = 1 THEN 0 ELSE 1 END,
              g.id
) c
WHERE a.rn = 1
