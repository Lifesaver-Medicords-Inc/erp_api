-- Grant the Admin position EVERY access code in the catalog (spec §3.2: "Admin
-- gets every module"). Part of the initial data a new database needs.
--
-- WHY THIS IS NEEDED
-- tbl_access_modules is only a CATALOG - the list of grantable codes. It is
-- seeded automatically on API startup (SeedAccessModules), which is why a fresh
-- database already has ~828 rows in it. A catalog entry is NOT a grant.
-- tbl_position_access is where grants live, and on a fresh database it is
-- EMPTY.
--
-- The consequence is a system nobody can get into. The Admin app logs in fine,
-- then asks for its position's access list, gets an empty one, and filters every
-- single sidebar entry out - the window opens completely blank. It is not
-- recoverable through the UI either: granting access requires the Access Control
-- screen, and reaching that screen requires access. Chicken and egg. This script
-- is the bootstrap that breaks it.
--
-- WHY ADMIN AND ONLY ADMIN
-- Do NOT extend this to seed every position. Grants are per position and
-- deliberately curated - on the live database Warehouse holds 70 codes, Sales
-- Representatives 105, Purchasing 39, and two positions hold none at all.
-- Seeding the whole catalog to all of them would hand Warehouse and Sales full
-- administrative rights. Every position other than Admin is granted through the
-- Access Control screen.
--
-- Idempotent: re-running grants nothing twice, and it picks up any codes added
-- to the catalog since it last ran, so it is safe to re-run after an upgrade
-- that introduces new codes.
--
-- Run AFTER the API has started at least once against the database (the catalog
-- rows must exist first, or this grants nothing).

SET NOCOUNT ON;

-- Resolve the Admin position rather than hardcoding an id: ids differ between
-- databases. Preference order:
--   1. whichever position the bootstrap user IT-WD-1 actually holds - that user
--      is the one that has to be able to log in, so its position is by
--      definition the one that needs everything
--   2. the position named 'Admin', for a database built before that user exists
DECLARE @position_id BIGINT;

SELECT TOP 1 @position_id = u.position_id
FROM tbl_setup_users u
WHERE u.employee_id = N'IT-WD-1'
  AND u.position_id IS NOT NULL;

IF @position_id IS NULL
BEGIN
    SELECT TOP 1 @position_id = p.id
    FROM tbl_position p
    WHERE p.name = N'Admin'
    ORDER BY p.id;
END

IF @position_id IS NULL
BEGIN
    RAISERROR(N'No Admin position found (looked for user IT-WD-1, then a position named ''Admin''). Nothing granted.', 16, 1);
    RETURN;
END

DECLARE @catalog_count INT;
SELECT @catalog_count = COUNT(DISTINCT code) FROM tbl_access_modules;

IF ISNULL(@catalog_count, 0) = 0
BEGIN
    RAISERROR(N'tbl_access_modules is empty - start the API once so SeedAccessModules populates the catalog, then re-run. Nothing granted.', 16, 1);
    RETURN;
END

DECLARE @before INT;
SELECT @before = COUNT(*) FROM tbl_position_access WHERE position_id = @position_id;

-- EXCEPT rather than NOT EXISTS: it also de-duplicates the catalog, which
-- carries repeated code rows (829 rows / 826 distinct on the live database).
INSERT INTO tbl_position_access (position_id, code)
SELECT @position_id, c.code
FROM (
    SELECT DISTINCT code FROM tbl_access_modules
    EXCEPT
    SELECT DISTINCT code FROM tbl_position_access WHERE position_id = @position_id
) c;

DECLARE @after INT;
SELECT @after = COUNT(*) FROM tbl_position_access WHERE position_id = @position_id;

PRINT N'Admin position id      : ' + CAST(@position_id AS NVARCHAR(20));
PRINT N'Catalog codes          : ' + CAST(@catalog_count AS NVARCHAR(20));
PRINT N'Grants before          : ' + CAST(@before AS NVARCHAR(20));
PRINT N'Grants after           : ' + CAST(@after AS NVARCHAR(20));
PRINT N'Newly granted          : ' + CAST(@after - @before AS NVARCHAR(20));

-- NOTE: the API caches this response in Redis under
-- model:PositionModel:conditions:...:preloads:Access. After running this,
-- restart the API (or clear that key) or the Admin app keeps being served the
-- cached, empty-access copy and still opens blank.
