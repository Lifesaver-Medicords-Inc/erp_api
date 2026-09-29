IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_bpi_finance_pending' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_bpi_finance_pending] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_bpi_finance_pending] AS
-- BPI Finance tab, PENDING (spec 4.1.7): the partner's quotations that have not become a sales
-- order yet (user instruction 2026-09-29: "pending for quotation"). Once an SO is made from any
-- copy of a quote, the quote leaves PENDING and its SO is listed under TRANSACTIONS
-- (vw_get_bpi_finance_transactions) instead, whatever the SO's status - so SO REF, which this
-- view carried until 2026-09-29, would always be blank here and is gone.
-- PENDING no longer feeds ACCOUNT BALANCE: that is the partner's receivable, worked out from
-- TRANSACTIONS (sales orders less payments). A quote is not owed.
-- Kept from 2026-09-26:
--   * PRICE is the quote's TOTAL AMOUNT DUE, the figure the printed quote and Opportunities show.
--   * One row per quote - its latest copy - with its own version's opportunity, else the newest.
--   * A quote whose opportunity STATUS is Abandoned or Lost is not pending (provisional standard:
--     the spec says only "ongoing").
WITH latest AS (
    -- One row per quote: its latest copy - highest version, then highest sub-version, then
    -- the FQ over the draft it was finalized from (the FQ is inserted after, so its id is
    -- higher). Keyed on the number without its Q#/FQ# prefix so a quote and its FQ are one
    -- quote. Every Edit session is its own row since versioning (2026-09-25); without this
    -- each one would show again here.
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
SELECT a.date,
    a.document_no + '.' + ISNULL(NULLIF(a.version_no, ''), '1') + '.' + ISNULL(NULLIF(a.sub_version_no, ''), '0') AS qoute_ref,
    ISNULL(a.total_amount_due, 0) AS total_price,
    o.stage,
    o.status,
    a.customer_id,
    d.id AS finance_pending_branch_id
FROM latest a
    OUTER APPLY (
        SELECT TOP 1 b.stage, b.status
          FROM tbl_trans_sales_opportunity b
         WHERE REPLACE(REPLACE(b.document_no, 'FQ#', ''), 'Q#', '') = a.bare_no
         ORDER BY CASE WHEN b.version_no = a.version_no THEN 0 ELSE 1 END, b.id DESC
    ) o
    LEFT JOIN (
        SELECT based_id, MIN(id) AS id
          FROM tbl_bpi_general
         GROUP BY based_id
    ) d ON a.customer_id = d.based_id
WHERE a.rn = 1
  AND ISNULL(o.status, '') NOT IN ('Abandoned', 'Lost')
  -- No SO made from any copy of this quote (the SO is made from the FQ, which need not be the
  -- latest copy). An SO whose quotation_id finds nothing is matched by the code it carries in
  -- document_no instead (FQ#2026-0004.1.0 - the quote's number with its version).
  AND NOT EXISTS (SELECT 1
                    FROM tbl_trans_sales_order s
                    LEFT JOIN tbl_trans_sales_quotation sq ON sq.id = s.quotation_id
                   WHERE REPLACE(REPLACE(sq.document_no, 'FQ#', ''), 'Q#', '') = a.bare_no
                      OR (sq.id IS NULL
                          AND REPLACE(REPLACE(s.document_no, 'FQ#', ''), 'Q#', '') LIKE a.bare_no + '.%'));
