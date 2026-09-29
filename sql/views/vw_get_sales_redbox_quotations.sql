IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_sales_redbox_quotations' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_sales_redbox_quotations] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_sales_redbox_quotations] AS
-- The quote cards of the Sales red box, first half (spec 9): one card per quote, from the moment
-- it is created until it becomes an SO. Added 2026-09-26 (user-reported: "I created a sales
-- quotation but not reflected on redbox"). The red box used to read GET /sales/quotation, which
--   * drops every PROJECT quotation (it keeps rows with no project name), so a project quote
--     never had a card at all;
--   * returns every copy of a quote - each version, each Edit sub-version and the FQ - so one
--     quote showed as several cards (Q#2026-0007 on test_fresh: four);
--   * carries every line and image of every quick quote, which the red box never reads.
-- Here: the latest copy of each quote (same order as Opportunities and the BPI Finance
-- PENDING list), aged from the FIRST copy's date - TIME ELAPSED counts from when the quote was
-- created (spec 9.1), and an Edit copy must not restart it.
--
-- A quote leaves once an SO is made from any of its copies (the SO card takes over; it is made
-- from the FQ, which need not be the latest copy), and - provisional standard, the same one the
-- BPI Finance PENDING list uses - once its opportunity STATUS is Abandoned or Lost, since the
-- card clears when the status reaches completion.
WITH copies AS (
    SELECT q.*,
        REPLACE(REPLACE(q.document_no, 'FQ#', ''), 'Q#', '') AS bare_no,
        ROW_NUMBER() OVER (
            PARTITION BY REPLACE(REPLACE(q.document_no, 'FQ#', ''), 'Q#', '')
            ORDER BY TRY_CAST(q.version_no AS int) DESC,
                     TRY_CAST(q.sub_version_no AS int) DESC,
                     q.id DESC
        ) AS rn
    FROM tbl_trans_sales_quotation q
)
SELECT a.id,
    a.document_no,
    a.version_no,
    a.sub_version_no,
    a.document_no + '.' + ISNULL(NULLIF(a.version_no, ''), '1') + '.' + ISNULL(NULLIF(a.sub_version_no, ''), '0') AS quote_code,
    a.is_finalized,
    ISNULL(a.project_name, '') AS project_name,
    a.customer_id,
    f.created_by,
    f.date
FROM copies a
    -- The first copy: who raised the quote, and when.
    CROSS APPLY (
        SELECT TOP 1 c.created_by, c.date
          FROM copies c
         WHERE c.bare_no = a.bare_no
         ORDER BY c.id
    ) f
    OUTER APPLY (
        SELECT TOP 1 b.status
          FROM tbl_trans_sales_opportunity b
         WHERE REPLACE(REPLACE(b.document_no, 'FQ#', ''), 'Q#', '') = a.bare_no
         ORDER BY CASE WHEN b.version_no = a.version_no THEN 0 ELSE 1 END, b.id DESC
    ) o
WHERE a.rn = 1
  AND ISNULL(o.status, '') NOT IN ('Abandoned', 'Lost')
  AND NOT EXISTS (
        SELECT 1
          FROM tbl_trans_sales_order s
          JOIN copies sc ON sc.id = s.quotation_id
         WHERE sc.bare_no = a.bare_no);
