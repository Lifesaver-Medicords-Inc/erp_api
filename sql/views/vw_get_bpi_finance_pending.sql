IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_bpi_finance_pending' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_bpi_finance_pending] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_bpi_finance_pending] AS
-- BPI Finance tab, PENDING (spec 4.1.7): the partner's ONGOING quotes and SOs. Its PRICE total
-- is the tab's ACCOUNT BALANCE (user decision 2026-09-17). Rewritten 2026-09-26 for a balance
-- that was wrong three ways:
--   * PRICE was gross_sales - the list-price gross before discounts and VAT on a quick quote
--     (Q#0002: 103,000 counted where the customer owes 87,864), and never filled in at all on a
--     project quote (0 on 3 of 3 on test_fresh). It is TOTAL AMOUNT DUE now, the figure the
--     printed quote and Opportunities show.
--   * The opportunity was joined by number alone, and the table holds one row per version (and,
--     on test_fresh, two for Q#0001 v1) - each extra row listed the quote again and counted it
--     again. One row per quote now: its own version's opportunity, else the newest.
--   * Everything was "pending". An SO that is finished - invoiced (the SO closes when its SI is
--     created, spec 5.4), CLOSED or CANCELLED - belongs to TRANSACTIONS
--     (vw_get_bpi_finance_transactions), and a quote whose opportunity STATUS is Abandoned or
--     Lost is not ongoing (provisional standard: the spec says only "ongoing").
-- SO REF (spec 4.1.7) was missing: the SO made from the quote, when there is one.
WITH latest AS (
    -- One row per quote: its latest copy - highest version, then highest sub-version, then
    -- the FQ over the draft it was finalized from (the FQ is inserted after, so its id is
    -- higher). Keyed on the number without its Q#/FQ# prefix so a quote and its FQ are one
    -- quote. Every Edit session is its own row since versioning (2026-09-25); without this
    -- each one would show again here and be counted again.
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
    so.so_ref,
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
    -- The SO made from any copy of this quote (it is made from the FQ, which need not be the
    -- latest copy), live ones first.
    OUTER APPLY (
        SELECT TOP 1 'SO#' + s.doc AS so_ref, s.doc, s.status
          FROM tbl_trans_sales_order s
          JOIN tbl_trans_sales_quotation sq ON sq.id = s.quotation_id
         WHERE REPLACE(REPLACE(sq.document_no, 'FQ#', ''), 'Q#', '') = a.bare_no
         ORDER BY CASE WHEN s.status = 'CANCELLED' THEN 1 ELSE 0 END, s.order_id DESC
    ) so
    LEFT JOIN (
        SELECT based_id, MIN(id) AS id
          FROM tbl_bpi_general
         GROUP BY based_id
    ) d ON a.customer_id = d.based_id
WHERE a.rn = 1
  AND ISNULL(o.status, '') NOT IN ('Abandoned', 'Lost')
  AND NOT (so.so_ref IS NOT NULL
           AND (ISNULL(so.status, '') IN ('CLOSED', 'CANCELLED')
                OR EXISTS (SELECT 1 FROM tbl_accounting_sales_invoice si
                            WHERE REPLACE(si.reference_doc_so, 'SO#', '') = so.doc)));
