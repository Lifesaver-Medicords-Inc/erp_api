-- STEP 01 of 01 - quotation numbers gain their year: Q#0026 -> Q#2026-0026, FQ#0004 -> FQ#2026-0004
-- Batch 2026-09-25_03 (quotation versioning). Run AFTER the API carrying the versioning change is in place.
-- A TRIAL RUN until you set @commit = 1; it must end by printing "committed".
--
-- Quotation versioning (user decision 2026-09-25)
-- ==============================================
-- A quotation's code is now Q#YYYY-nnnn.version.sub-version - Q#2026-0001.2.1 - with the year
-- the quote was raised and a count that restarts every January (spec §2). The API numbers new
-- quotes that way already; this brings the ones that exist into line:
--
--   tbl_trans_sales_quotation.document_no   Q#0026 -> Q#2026-0026, FQ#0004 -> FQ#2026-0004
--                                            (year = the quote's own DATE; version and
--                                             sub-version are separate columns and do not change)
--   tbl_trans_sales_opportunity.document_no  the same rename, so its join to the quote still holds
--   tbl_trans_sales_order.document_no        the source FQ's FULL code - FQ#2026-0004.1.0 - in
--                                            place of the bare number "0004", which once versions
--                                            exist names several FQs. quotation_id stays the link.
--
-- NOT changed: the audit table (tbl_trans_sales_quotation_at) keeps the numbers as they were
-- when each change was made - it is history. Sales orders made from a DRAFT (which the API now
-- refuses) are listed, not altered.
--
-- Safe to run again: a number that already has its year is left alone.
--
-- HOW TO RUN: open in SSMS (or sqlcmd -f 65001 -b -i <this file>).
-- After committing, restart the API or clear its Redis keys - quotation lists are cached.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;   -- 1 = keep the changes, 0 = trial run

IF OBJECT_ID('tempdb..#renumber') IS NOT NULL DROP TABLE #renumber;

-- ---- the mapping ---------------------------------------------------------------------
SELECT q.id,
       q.document_no AS old_no,
       LEFT(q.document_no, CHARINDEX('#', q.document_no))                       -- "Q#" or "FQ#"
         + CAST(YEAR(TRY_CONVERT(datetime, q.[date])) AS varchar(4)) + '-'
         + CASE WHEN TRY_CONVERT(int, SUBSTRING(q.document_no, CHARINDEX('#', q.document_no) + 1, 20)) > 9999
                THEN CAST(TRY_CONVERT(int, SUBSTRING(q.document_no, CHARINDEX('#', q.document_no) + 1, 20)) AS varchar(10))
                ELSE RIGHT('0000' + CAST(TRY_CONVERT(int, SUBSTRING(q.document_no, CHARINDEX('#', q.document_no) + 1, 20)) AS varchar(10)), 4)
           END AS new_no,
       q.version_no,
       q.sub_version_no
  INTO #renumber
  FROM tbl_trans_sales_quotation q
 WHERE CHARINDEX('#', q.document_no) > 0
   AND CHARINDEX('-', q.document_no) = 0;          -- not yet renamed

-- ---- checks: every one must pass, or nothing is written ----------------------------------
IF EXISTS (SELECT 1 FROM #renumber WHERE new_no IS NULL
           OR TRY_CONVERT(int, SUBSTRING(old_no, CHARINDEX('#', old_no) + 1, 20)) IS NULL)
BEGIN
    SELECT 'STOPPED: no year or no number' AS problem, r.id, r.old_no, q.[date]
      FROM #renumber r JOIN tbl_trans_sales_quotation q ON q.id = r.id
     WHERE r.new_no IS NULL OR TRY_CONVERT(int, SUBSTRING(r.old_no, CHARINDEX('#', r.old_no) + 1, 20)) IS NULL;
    PRINT 'Some quotations have no readable DATE or number (listed above). Nothing was changed.';
    RETURN;
END

IF EXISTS (SELECT 1 FROM #renumber r
             JOIN tbl_trans_sales_quotation q
               ON q.document_no = r.new_no AND q.version_no = r.version_no AND q.sub_version_no = r.sub_version_no)
    OR EXISTS (SELECT new_no, version_no, sub_version_no FROM #renumber
                GROUP BY new_no, version_no, sub_version_no HAVING COUNT(*) > 1)
BEGIN
    SELECT 'STOPPED: the new number is already taken' AS problem, r.id, r.old_no, r.new_no, r.version_no, r.sub_version_no
      FROM #renumber r
     WHERE EXISTS (SELECT 1 FROM tbl_trans_sales_quotation q
                    WHERE q.document_no = r.new_no AND q.version_no = r.version_no AND q.sub_version_no = r.sub_version_no)
        OR EXISTS (SELECT 1 FROM #renumber r2
                    WHERE r2.id <> r.id AND r2.new_no = r.new_no AND r2.version_no = r.version_no AND r2.sub_version_no = r.sub_version_no);
    PRINT 'Renaming would give two quotations the same code (listed above). Nothing was changed.';
    RETURN;
END

BEGIN TRANSACTION;

SELECT 'to be renamed' AS state, id, old_no,
       new_no + '.' + ISNULL(NULLIF(version_no, ''), '1') + '.' + ISNULL(NULLIF(sub_version_no, ''), '0') AS new_code
  FROM #renumber
 ORDER BY id;

UPDATE q SET q.document_no = r.new_no
  FROM tbl_trans_sales_quotation q
  JOIN #renumber r ON r.id = q.id;

-- Opportunities follow their quote (same document_no and version, as the view joins them).
UPDATE o SET o.document_no = r.new_no
  FROM tbl_trans_sales_opportunity o
  JOIN (SELECT DISTINCT old_no, new_no FROM #renumber) r ON r.old_no = o.document_no;

-- Sales orders record WHICH FQ they came from, in full.
UPDATE so SET so.document_no = q.document_no + '.' + ISNULL(NULLIF(q.version_no, ''), '1') + '.' + ISNULL(NULLIF(q.sub_version_no, ''), '0')
  FROM tbl_trans_sales_order so
  JOIN tbl_trans_sales_quotation q ON q.id = so.quotation_id;

SELECT 'sales orders now read' AS state, so.order_id, 'SO#' + so.doc AS so_no, so.document_no AS from_quote,
       CASE WHEN q.is_finalized = 1 THEN '' ELSE 'MADE FROM A DRAFT - the API now refuses this; review it' END AS note
  FROM tbl_trans_sales_order so
  LEFT JOIN tbl_trans_sales_quotation q ON q.id = so.quotation_id
 ORDER BY so.order_id;

SELECT 'after' AS state,
       (SELECT COUNT(*) FROM tbl_trans_sales_quotation WHERE CHARINDEX('-', document_no) = 0) AS quotations_without_year,
       (SELECT COUNT(*) FROM tbl_trans_sales_quotation) AS quotations_total;

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT 'committed - restart the API or clear its Redis keys; quotation lists are cached';
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT 'rolled back - set @commit = 1 to write';
END

IF OBJECT_ID('tempdb..#renumber') IS NOT NULL DROP TABLE #renumber;
