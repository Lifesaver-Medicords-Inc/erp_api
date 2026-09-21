-- Clear the grants of positions whose access predates the access-level workbook
-- =============================================================================
-- SeedPositionAccess (initializers/seed_position_access.go) deliberately leaves alone any
-- position that can already open a screen, so a grant somebody removed through Admin's
-- Access Control is never handed back by a redeploy. That protects curated positions - but
-- it also preserved four positions whose grants came from earlier experimentation and do
-- not match the workbook at all. Warehouse, for one, held no Inventory App codes whatever,
-- so with navigation now filtered by access its sidebar would come up empty in its own app.
--
-- This empties those four so the seeder treats them as un-set-up and fills them from the
-- workbook on the next API start (or the next run of TestSeedPositionAccess).
--
-- Set @commit = 1 to write. With @commit = 0 it reports and rolls back.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;

-- Named, not pattern-matched: this deletes access, so the list is explicit.
DECLARE @positions TABLE (name nvarchar(200));
INSERT INTO @positions (name) VALUES
    ('Warehouse'),
    ('Purschasing'),     -- spelled this way on some databases
    ('Purchaser'),
    ('Purchasing'),
    ('Sales Representatives'),
    ('Accounts Receivables');

BEGIN TRANSACTION;

SELECT 'before' AS state, p.name, COUNT(a.id) AS grants
FROM tbl_position p
LEFT JOIN tbl_position_access a ON a.position_id = p.id
WHERE p.name IN (SELECT name FROM @positions)
GROUP BY p.name
ORDER BY p.name;

DELETE a
FROM tbl_position_access a
JOIN tbl_position p ON p.id = a.position_id
WHERE p.name IN (SELECT name FROM @positions);

PRINT CONCAT('grants cleared: ', @@ROWCOUNT);

-- Nothing outside the named positions may have been touched.
IF EXISTS (
    SELECT 1 FROM tbl_position p
    JOIN tbl_position_access a ON a.position_id = p.id
    WHERE p.name IN (SELECT name FROM @positions))
BEGIN
    RAISERROR('a named position still holds grants after the delete', 16, 1);
    ROLLBACK TRANSACTION;
    RETURN;
END

SELECT 'after' AS state, p.name, COUNT(a.id) AS grants
FROM tbl_position p
LEFT JOIN tbl_position_access a ON a.position_id = p.id
GROUP BY p.name
ORDER BY p.name;

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT 'committed - restart the API (or run TestSeedPositionAccess) to refill them';
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT 'rolled back - set @commit = 1 to write';
END
