-- ============================================================
-- Give BPI branches back to the owner they had before a later EDIT took
-- them over.
--
-- WHY: until the fix in UpdateBpiGeneral (services/bpi_services/
-- bpi_general.services.go, 2026-09-11), saving an edit wrote the EDITOR's
-- name into tbl_bpi_general.sales_id - so whoever saved a partner became its
-- owner, and it appeared in their quotation customer picker. Spec 4.1.10 and
-- 14.168: editing a record never changes its owner.
--
-- RULE 1 - branches created in the app. Their first audit row
-- (z_tbl_bpi_general_at, lowest id) is the creation and holds the owner they
-- were created with. Older audit rows hold the owner's EMPLOYEE ID (from
-- before owners were stored as names), so that is mapped to the user's full
-- name. A branch whose current owner differs is restored to it.
--
-- RULE 2 - main branches loaded by the QERP import (notes 'QERP entity
-- code: ...'). The import writes no audit row, so their first audit row is an
-- EDIT. If that edit stamped its own author as the owner, the owner before it
-- is the company's: the import writes the same owner to the company and its
-- main branch.
--
-- Imported CHILD branches stamped the same way are only LISTED at the end:
-- nothing records who owned them before, so they need a person to decide.
--
-- There is no reassignment feature yet (spec 3.3 / 4.1.10 - unbuilt), so every
-- such difference was made by an edit. Once reassignment exists, do NOT re-run
-- this without excluding branches that were reassigned on purpose.
--
-- AFTER running it on a database the API is serving, clear the API cache -
-- GET <API base>/admin/clear_all, the same call the Engineering app makes
-- (ApiEndPoints.CLEAR_CACHE) - or the old owners may keep showing until the
-- cache expires.
--
-- REVERSIBLE: every change is recorded in zz_bpi_owner_restore_backup.
-- IDEMPOTENT: a second run finds nothing to restore. SQL Server 2012.
-- ============================================================
SET NOCOUNT ON;
SET XACT_ABORT ON;

IF OBJECT_ID(N'zz_bpi_owner_restore_backup', N'U') IS NULL
    CREATE TABLE zz_bpi_owner_restore_backup (
        general_id    BIGINT        NOT NULL,
        owner_before  NVARCHAR(200) NULL,
        owner_after   NVARCHAR(200) NULL,
        restored_at   DATETIME      NOT NULL DEFAULT GETDATE()
    );
GO

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @restored INT;
DECLARE @fix TABLE (general_id BIGINT PRIMARY KEY, owner_before NVARCHAR(200), owner_after NVARCHAR(200), rule_no INT);

-- RULE 1: app-created branches -> the owner they were created with
WITH firsts AS (
    SELECT a.ref_id, a.sales_id,
           ROW_NUMBER() OVER (PARTITION BY a.ref_id ORDER BY a.id) AS rn
    FROM z_tbl_bpi_general_at a
),
creators AS (
    SELECT f.ref_id,
           LTRIM(RTRIM(ISNULL(LTRIM(RTRIM(u.first_name)) + N' ' + LTRIM(RTRIM(ISNULL(u.last_name, N''))),
                              ISNULL(f.sales_id, N'')))) AS creator
    FROM firsts f
    LEFT JOIN tbl_setup_users u ON u.employee_id = f.sales_id
    WHERE f.rn = 1
)
INSERT INTO @fix (general_id, owner_before, owner_after, rule_no)
SELECT g.id, g.sales_id, c.creator, 1
FROM tbl_bpi_general g
JOIN creators c ON c.ref_id = g.id
WHERE ISNULL(g.notes, N'') NOT LIKE N'QERP entity code:%'
  AND c.creator <> N''
  AND UPPER(LTRIM(RTRIM(ISNULL(g.sales_id, N'')))) <> UPPER(c.creator);

-- RULE 2: imported main branches an edit stamped with its own author -> the company's owner
WITH firsts AS (
    SELECT a.ref_id, a.AT_USER_ID,
           ROW_NUMBER() OVER (PARTITION BY a.ref_id ORDER BY a.id) AS rn
    FROM z_tbl_bpi_general_at a
)
INSERT INTO @fix (general_id, owner_before, owner_after, rule_no)
SELECT g.id, g.sales_id, LTRIM(RTRIM(b.sales_id)), 2
FROM tbl_bpi_general g
JOIN tbl_bpi b ON b.id = g.based_id
JOIN firsts f ON f.ref_id = g.id AND f.rn = 1
JOIN tbl_setup_users u ON CAST(u.id AS NVARCHAR(20)) = f.AT_USER_ID
WHERE g.notes LIKE N'QERP entity code:%'
  AND g.is_main = 1
  AND LTRIM(RTRIM(ISNULL(b.sales_id, N''))) <> N''
  AND UPPER(LTRIM(RTRIM(ISNULL(g.sales_id, N'')))) =
      UPPER(LTRIM(RTRIM(LTRIM(RTRIM(u.first_name)) + N' ' + LTRIM(RTRIM(ISNULL(u.last_name, N''))))))
  AND UPPER(LTRIM(RTRIM(ISNULL(g.sales_id, N'')))) <> UPPER(LTRIM(RTRIM(b.sales_id)))
  AND g.id NOT IN (SELECT general_id FROM @fix);

BEGIN TRAN;

INSERT INTO zz_bpi_owner_restore_backup (general_id, owner_before, owner_after)
SELECT general_id, owner_before, owner_after FROM @fix;

UPDATE g
SET sales_id = f.owner_after
FROM tbl_bpi_general g
JOIN @fix f ON f.general_id = g.id;
SET @restored = @@ROWCOUNT;

COMMIT;

PRINT N'branches given back to their previous owner: ' + CAST(@restored AS NVARCHAR(20));

SELECT f.rule_no, f.general_id, LEFT(b.name, 40) AS company, LEFT(g.branch_name, 40) AS branch,
       f.owner_before, f.owner_after
FROM @fix f
JOIN tbl_bpi_general g ON g.id = f.general_id
JOIN tbl_bpi b ON b.id = g.based_id
ORDER BY f.general_id;

-- Imported child branches an edit stamped with its own author: decide by hand.
WITH firsts AS (
    SELECT a.ref_id, a.AT_USER_ID, a.AT_DATE,
           ROW_NUMBER() OVER (PARTITION BY a.ref_id ORDER BY a.id) AS rn
    FROM z_tbl_bpi_general_at a
)
SELECT g.id AS general_id, LEFT(b.name, 40) AS company, LEFT(g.branch_name, 40) AS branch,
       g.sales_id AS owner_now, b.sales_id AS company_owner, f.AT_DATE AS first_edit
FROM tbl_bpi_general g
JOIN tbl_bpi b ON b.id = g.based_id
JOIN firsts f ON f.ref_id = g.id AND f.rn = 1
JOIN tbl_setup_users u ON CAST(u.id AS NVARCHAR(20)) = f.AT_USER_ID
WHERE g.notes LIKE N'QERP entity code:%'
  AND g.is_main = 0
  AND UPPER(LTRIM(RTRIM(ISNULL(g.sales_id, N'')))) =
      UPPER(LTRIM(RTRIM(LTRIM(RTRIM(u.first_name)) + N' ' + LTRIM(RTRIM(ISNULL(u.last_name, N''))))));
GO
