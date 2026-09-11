-- ============================================================
-- Replace the legacy department label "Accounting" on user accounts
-- with the specific department. Spec 3.2: "Accounting" MUST NOT
-- appear anywhere - it is A/R, A/P, or A/R-A/P Cashier; pick the
-- specific one.
--
-- The user's position decides which:
--     Accounts Receivables  -> A/R
--     Accounts Payable      -> A/P
-- A user in "Accounting" with any other position is NOT guessed at:
-- it is listed at the end, unchanged, for a person to decide.
--
-- Only tbl_setup_users.department changes. Left exactly as they are:
--   - the employee id (e.g. "Accounting--10004") - it is the user's
--     LOGIN, so changing it is a separate decision;
--   - audit rows (z_tbl_setup_users_at);
--   - department values already written onto documents (e.g. an
--     Item Request's REQUESTING DEPT.) - those are history.
--
-- REVERSIBLE: every change is recorded in zz_user_department_backup.
-- IDEMPOTENT: a second run finds nothing to change.
-- SQL Server 2012 compatible.
-- ============================================================
SET NOCOUNT ON;
SET XACT_ABORT ON;

IF OBJECT_ID(N'zz_user_department_backup', N'U') IS NULL
    CREATE TABLE zz_user_department_backup (
        user_id         BIGINT        NOT NULL,
        employee_id     NVARCHAR(100) NULL,
        old_department  NVARCHAR(100) NULL,
        new_department  NVARCHAR(100) NOT NULL,
        changed_at      DATETIME      NOT NULL DEFAULT GETDATE()
    );
GO

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @changed INT;
DECLARE @map TABLE (position_name NVARCHAR(200) PRIMARY KEY, department NVARCHAR(100) NOT NULL);
INSERT INTO @map (position_name, department) VALUES
    (N'Accounts Receivables', N'A/R'),
    (N'Accounts Payable',     N'A/P');

BEGIN TRAN;

INSERT INTO zz_user_department_backup (user_id, employee_id, old_department, new_department)
SELECT u.id, u.employee_id, u.department, m.department
FROM tbl_setup_users u
JOIN tbl_position p ON p.id = u.position_id
JOIN @map m ON m.position_name = LTRIM(RTRIM(p.name))
WHERE LTRIM(RTRIM(u.department)) = N'Accounting';

UPDATE u
SET department = m.department
FROM tbl_setup_users u
JOIN tbl_position p ON p.id = u.position_id
JOIN @map m ON m.position_name = LTRIM(RTRIM(p.name))
WHERE LTRIM(RTRIM(u.department)) = N'Accounting';
SET @changed = @@ROWCOUNT;

COMMIT;

PRINT N'users moved off "Accounting" : ' + CAST(@changed AS NVARCHAR(20));

-- Anything still here needs a person to pick A/R, A/P or A/R-A/P Cashier.
SELECT u.id, u.employee_id, u.first_name + N' ' + u.last_name AS name, p.name AS position,
       u.department AS still_accounting
FROM tbl_setup_users u
LEFT JOIN tbl_position p ON p.id = u.position_id
WHERE LTRIM(RTRIM(u.department)) = N'Accounting';
GO
