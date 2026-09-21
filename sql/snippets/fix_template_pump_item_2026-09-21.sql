-- Project templates: repoint rows whose item no longer exists
-- ============================================================
-- A template child row stores the id of the catalogue item it was built from. The model
-- picker uses that item only as an anchor - it lists every item sharing the anchor's item
-- name - so the row shows the model list for its component.
--
-- On a database rebuilt from scratch the catalogue was renumbered (items start at 1245 on
-- Lightspeed_test_fresh), but the template rows kept the old ids. Every template's PUMP row
-- still pointed at item 1, which no longer exists, so the picker had no item name to scope
-- by and came up "No models found" (user-reported 2026-09-21).
--
-- This repoints any dangling row to an item whose item_name matches the component name, and
-- reports anything it cannot match rather than guessing. Which pump it lands on does not
-- matter and is never shown: it only names the list.
--
-- Set @commit = 1 to write. With @commit = 0 it prints what it would do and rolls back.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;

BEGIN TRANSACTION;

-- The dangling rows, with the item this would point each one at.
IF OBJECT_ID('tempdb..#fix') IS NOT NULL DROP TABLE #fix;

SELECT c.id,
       c.item_id AS old_item_id,
       c.components,
       t.template_name,
       (SELECT MIN(i.id)
          FROM tbl_setup_item i
          JOIN tbl_setup_item_name n ON n.id = i.item_name_id
         WHERE n.name = LTRIM(RTRIM(c.components))) AS new_item_id
INTO #fix
FROM tbl_trans_sales_project_template_child c
LEFT JOIN tbl_setup_item existing ON existing.id = c.item_id
LEFT JOIN tbl_trans_sales_project_template t ON t.template_id = c.parent_id
WHERE existing.id IS NULL;

SELECT 'to fix' AS state, template_name, id AS template_row, components, old_item_id, new_item_id
FROM #fix ORDER BY id;

-- A component whose name matches no item is left alone and named here: it needs a decision,
-- not a guess.
IF EXISTS (SELECT 1 FROM #fix WHERE new_item_id IS NULL)
BEGIN
    SELECT 'NO MATCHING ITEM - left as is' AS state, template_name, id AS template_row, components, old_item_id
    FROM #fix WHERE new_item_id IS NULL ORDER BY id;
END

UPDATE c
   SET c.item_id = f.new_item_id
  FROM tbl_trans_sales_project_template_child c
  JOIN #fix f ON f.id = c.id
 WHERE f.new_item_id IS NOT NULL;

PRINT CONCAT('rows repointed: ', @@ROWCOUNT);

-- Nothing may still be dangling after the update, except the rows named above.
IF EXISTS (
    SELECT 1
      FROM tbl_trans_sales_project_template_child c
      LEFT JOIN tbl_setup_item i ON i.id = c.item_id
     WHERE i.id IS NULL
       AND c.id NOT IN (SELECT id FROM #fix WHERE new_item_id IS NULL))
BEGIN
    RAISERROR('a template row is still pointing at a missing item', 16, 1);
    ROLLBACK TRANSACTION;
    RETURN;
END

SELECT 'after' AS state, t.template_name, c.id AS template_row, c.components, c.item_id,
       i.item_name + ' | ' + i.item_model AS now_points_at
FROM tbl_trans_sales_project_template_child c
JOIN #fix f ON f.id = c.id
LEFT JOIN vw_items i ON i.id = c.item_id
LEFT JOIN tbl_trans_sales_project_template t ON t.template_id = c.parent_id
ORDER BY c.id;

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT 'committed';
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT 'rolled back - set @commit = 1 to write';
END

-- #fix is created inside the transaction, so a rollback has already taken it with it.
IF OBJECT_ID('tempdb..#fix') IS NOT NULL DROP TABLE #fix;
