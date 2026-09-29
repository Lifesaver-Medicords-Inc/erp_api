-- STEP 01 of 01 - repairs data damaged by the project quotation save bugs (is_project, orphaned tab rows)
-- Bugfix round of 2026-09-25. Run after the scripts in 2026-09-25_01_controllers_and_project_quote.
-- A TRIAL RUN until you set @commit = 1; it must end by printing "committed".

-- Project quotation data damaged by save bugs (fixed 2026-09-24/25)
-- ==============================================================
-- Two things the project quotation's save did to data that is already stored. The code no
-- longer does either; this puts right what it left behind.
--
-- 1. is_project was set to NULL/0 on every project quotation it saved.
--    The save compared is_project but never sent it, and read "not sent" as "cleared". Every
--    project quote on the rehearsal database reads is_project = 0. A project quotation is one
--    with a project name - the rule the project list, the engineering list view and the API's
--    detail endpoint all use - so that is what restores it.
--
-- 2. Removing a tab (item set) deleted the tab row and left its children behind.
--    Content, its Size Up and Final Selection rows, Advanced Conditions and wiring rows whose
--    item set no longer exists. Invisible in the app, but still read by anything that reads
--    those tables whole. Change History for removed tabs is KEPT - it is the audit trail.
--
-- NOT repaired, deliberately: final_ref_no and discounted_amount are blank on every quotation
-- on the rehearsal database, quick quotes included, so they were not populated in the first
-- place and there is nothing to restore them from.
--
-- Items are never deleted here. An item can hold a stock reservation, and releasing one is
-- the application's job (releaseProjectItemReservation). If orphaned items exist the script
-- stops and rolls back instead.
--
-- HOW TO RUN: sqlcmd -S <server> -d <database> -b -i <this file>  (or open it in SSMS).
-- With @commit = 0 everything is worked out and reported, then rolled back.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;   -- 1 = keep the changes, 0 = trial run

BEGIN TRANSACTION;

-- ---- guard: orphaned items need the app, not this script -------------------------------
DECLARE @orphan_items int = (
    SELECT COUNT(*) FROM tbl_trans_sales_project_items i
     WHERE NOT EXISTS (SELECT 1 FROM tbl_trans_sales_project_item_set s WHERE s.item_set_id = i.based_id));

IF @orphan_items > 0
BEGIN
    SELECT 'STOPPED: orphaned project items exist' AS state, @orphan_items AS rows_;
    PRINT 'Orphaned items may hold stock reservations; nothing was changed. Resolve them in the app.';
    ROLLBACK TRANSACTION;
    RETURN;
END

-- ---- 1. is_project --------------------------------------------------------------------
SELECT 'is_project restored' AS state, id, document_no, project_name
  FROM tbl_trans_sales_quotation
 WHERE ISNULL(project_name, '') <> '' AND ISNULL(is_project, 0) = 0
 ORDER BY id;

UPDATE tbl_trans_sales_quotation
   SET is_project = 1
 WHERE ISNULL(project_name, '') <> '' AND ISNULL(is_project, 0) = 0;

-- ---- 2. orphans of removed tabs, children first ------------------------------------------
IF OBJECT_ID('tempdb..#orphan_content') IS NOT NULL DROP TABLE #orphan_content;

SELECT c.content_id, c.based_id
  INTO #orphan_content
  FROM tbl_trans_sales_project_content c
 WHERE NOT EXISTS (SELECT 1 FROM tbl_trans_sales_project_item_set s WHERE s.item_set_id = c.based_id);

SELECT 'orphans to remove' AS state, 'size up' AS what, COUNT(*) AS rows_
  FROM tbl_trans_sales_project_size_up u JOIN #orphan_content o ON o.content_id = u.sales_project_content_id
UNION ALL
SELECT 'orphans to remove', 'final selection', COUNT(*)
  FROM tbl_trans_sales_project_content_final f JOIN #orphan_content o ON o.content_id = f.sales_project_content_id
UNION ALL
SELECT 'orphans to remove', 'content', COUNT(*) FROM #orphan_content
UNION ALL
SELECT 'orphans to remove', 'advanced conditions', COUNT(*)
  FROM tbl_trans_sales_project_content_advanced_configuration a
 WHERE NOT EXISTS (SELECT 1 FROM tbl_trans_sales_project_item_set s WHERE s.item_set_id = a.based_id)
UNION ALL
SELECT 'orphans to remove', 'wiring', COUNT(*)
  FROM tbl_trans_sales_project_wiring w
 WHERE NOT EXISTS (SELECT 1 FROM tbl_trans_sales_project_item_set s WHERE s.item_set_id = w.based_id);

DELETE u FROM tbl_trans_sales_project_size_up u
  JOIN #orphan_content o ON o.content_id = u.sales_project_content_id;

DELETE f FROM tbl_trans_sales_project_content_final f
  JOIN #orphan_content o ON o.content_id = f.sales_project_content_id;

DELETE c FROM tbl_trans_sales_project_content c
  JOIN #orphan_content o ON o.content_id = c.content_id;

DELETE a FROM tbl_trans_sales_project_content_advanced_configuration a
 WHERE NOT EXISTS (SELECT 1 FROM tbl_trans_sales_project_item_set s WHERE s.item_set_id = a.based_id);

DELETE w FROM tbl_trans_sales_project_wiring w
 WHERE NOT EXISTS (SELECT 1 FROM tbl_trans_sales_project_item_set s WHERE s.item_set_id = w.based_id);

-- ---- after ----------------------------------------------------------------------------
SELECT 'after' AS state,
       (SELECT COUNT(*) FROM tbl_trans_sales_quotation WHERE ISNULL(project_name, '') <> '' AND ISNULL(is_project, 0) = 0) AS projects_without_flag,
       (SELECT COUNT(*) FROM tbl_trans_sales_project_content c WHERE NOT EXISTS (SELECT 1 FROM tbl_trans_sales_project_item_set s WHERE s.item_set_id = c.based_id)) AS orphan_content,
       (SELECT COUNT(*) FROM tbl_trans_sales_project_content_advanced_configuration a WHERE NOT EXISTS (SELECT 1 FROM tbl_trans_sales_project_item_set s WHERE s.item_set_id = a.based_id)) AS orphan_conditions,
       (SELECT COUNT(*) FROM tbl_trans_sales_project_wiring w WHERE NOT EXISTS (SELECT 1 FROM tbl_trans_sales_project_item_set s WHERE s.item_set_id = w.based_id)) AS orphan_wiring;

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

IF OBJECT_ID('tempdb..#orphan_content') IS NOT NULL DROP TABLE #orphan_content;
