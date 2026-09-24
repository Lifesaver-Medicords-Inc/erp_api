-- Project template rows saved with level 0
-- ========================================
-- There is no level 0 in a project template: level 1 is the top, which is what every row
-- saved before this holds and what the project quotation assumes when it applies a
-- template. The Template Setup screen used to write 0 for a parent, and re-picking the item
-- on an existing row ran through that same path - so a row that was a child came back as
-- level 0 after somebody edited it.
--
-- On Lightspeed_test_fresh that is CPS's DISCHARGE RUBBER BELLOW and SUCTION RUBBER BELLOW,
-- both sitting between level-3 siblings, and both level 3 on the rehearsal database which
-- has not been edited since.
--
-- The repair gives each level-0 row the level of THE ROW ABOVE IT in the same template -
-- its sibling context, which is the only evidence in the data of where it belongs - and 1
-- when it is the first row of its template. Nothing else is touched. Every change is listed
-- before it is kept, so the result can be checked against what the screen should show.
--
-- The screen no longer produces these (a parent is written as level 1, and re-picking an
-- item keeps the row's level), so this is a one-off clean-up of what is already stored.
--
-- HOW TO RUN: sqlcmd -S <server> -d <database> -b -i <this file>  (or open it in SSMS).
-- With @commit = 0 everything is worked out and reported, then rolled back.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;   -- 1 = keep the changes, 0 = trial run

IF OBJECT_ID('tempdb..#fix') IS NOT NULL DROP TABLE #fix;

BEGIN TRANSACTION;

-- The row above, within the same template, in id order - which is the order the screen
-- lists them in.
SELECT c.id,
       c.parent_id,
       c.components,
       c.level AS old_level,
       ISNULL((SELECT TOP 1 p.level
                 FROM tbl_trans_sales_project_template_child p
                WHERE p.parent_id = c.parent_id
                  AND p.id < c.id
                  AND p.level > 0
                ORDER BY p.id DESC), 1) AS new_level
  INTO #fix
  FROM tbl_trans_sales_project_template_child c
 WHERE c.level = 0;

SELECT 'to be corrected' AS state, f.id, t.template_name, f.components,
       f.old_level, f.new_level
  FROM #fix f
  LEFT JOIN tbl_trans_sales_project_template t ON t.template_id = f.parent_id
 ORDER BY f.parent_id, f.id;

UPDATE c
   SET c.level = f.new_level
  FROM tbl_trans_sales_project_template_child c
  JOIN #fix f ON f.id = c.id;

SELECT 'rows corrected' AS state, COUNT(*) AS rows_ FROM #fix;
SELECT 'levels now in use' AS state, level, COUNT(*) AS rows_
  FROM tbl_trans_sales_project_template_child
 GROUP BY level ORDER BY level;

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT 'committed - restart the API or clear its Redis keys, template lists are cached';
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT 'rolled back - set @commit = 1 to write';
END

IF OBJECT_ID('tempdb..#fix') IS NOT NULL DROP TABLE #fix;
