-- ============================================================
-- Rename a sales executive on the Business Partners they own.
--
-- WHY: BPI ownership is matched by NAME - a partner's sales_id holds its
-- owner's first name + " " + last name. When a user is renamed in the Admin
-- app, their partners must be renamed in the same step; otherwise nobody owns
-- them: the renamed user can no longer edit them or see them in their
-- quotation customer picker. (First case: SAL-SR-4 "J. CESTONA" renamed to
-- "Julie CESTONA", 2026-09-11.)
--
-- HOW TO RUN (sqlcmd variables; in SSMS switch on Query > SQLCMD Mode):
--     sqlcmd -S <server> -U <user> -P <pw> -d <database> -b -I ^
--            -v OldName="J. CESTONA" NewName="Julie CESTONA" -i rename_bpi_owner.sql
-- From PowerShell, put --% straight after sqlcmd (sqlcmd --% -S ... -v ...) so
-- the quoted names reach sqlcmd intact.
-- NewName must be EXACTLY the renamed user's first name + " " + last name
-- (letter case is ignored); the script refuses to run if no such user exists.
--
-- Changes tbl_bpi.sales_id and tbl_bpi_general.sales_id only. History rows
-- keep the name they were written with.
--
-- AFTER running it on a database the API is serving, clear the API cache -
-- GET <API base>/admin/clear_all (the Engineering app's CLEAR_CACHE call) - or
-- the old owner may keep showing until the cache expires.
--
-- REVERSIBLE: every change is recorded in zz_bpi_owner_rename_backup.
-- IDEMPOTENT: a second run finds nothing left under the old name.
-- SQL Server 2012 compatible.
-- ============================================================
SET NOCOUNT ON;
SET XACT_ABORT ON;

IF OBJECT_ID(N'zz_bpi_owner_rename_backup', N'U') IS NULL
    CREATE TABLE zz_bpi_owner_rename_backup (
        table_name    NVARCHAR(40)  NOT NULL,
        row_id        BIGINT        NOT NULL,
        owner_before  NVARCHAR(200) NULL,
        owner_after   NVARCHAR(200) NULL,
        renamed_at    DATETIME      NOT NULL DEFAULT GETDATE()
    );
GO

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @old NVARCHAR(200) = LTRIM(RTRIM(N'$(OldName)'));
DECLARE @new NVARCHAR(200) = LTRIM(RTRIM(N'$(NewName)'));
DECLARE @companies INT, @branches INT;

IF @old = N'' OR @new = N''
BEGIN
    RAISERROR(N'Give both names: -v OldName="..." NewName="..."', 16, 1);
    RETURN;
END

IF NOT EXISTS (SELECT 1 FROM tbl_setup_users u
               WHERE UPPER(LTRIM(RTRIM(LTRIM(RTRIM(u.first_name)) + N' ' + LTRIM(RTRIM(ISNULL(u.last_name, N'')))))) = UPPER(@new))
BEGIN
    RAISERROR(N'No user is named "%s". Rename the user first (Admin app), then run this with that exact name.', 16, 1, @new);
    RETURN;
END

BEGIN TRAN;

INSERT INTO zz_bpi_owner_rename_backup (table_name, row_id, owner_before, owner_after)
SELECT N'tbl_bpi', id, sales_id, @new FROM tbl_bpi WHERE UPPER(LTRIM(RTRIM(sales_id))) = UPPER(@old);
UPDATE tbl_bpi SET sales_id = @new WHERE UPPER(LTRIM(RTRIM(sales_id))) = UPPER(@old);
SET @companies = @@ROWCOUNT;

INSERT INTO zz_bpi_owner_rename_backup (table_name, row_id, owner_before, owner_after)
SELECT N'tbl_bpi_general', id, sales_id, @new FROM tbl_bpi_general WHERE UPPER(LTRIM(RTRIM(sales_id))) = UPPER(@old);
UPDATE tbl_bpi_general SET sales_id = @new WHERE UPPER(LTRIM(RTRIM(sales_id))) = UPPER(@old);
SET @branches = @@ROWCOUNT;

COMMIT;

PRINT N'"' + @old + N'" -> "' + @new + N'":  companies ' + CAST(@companies AS NVARCHAR(20))
    + N', branches ' + CAST(@branches AS NVARCHAR(20));
GO
