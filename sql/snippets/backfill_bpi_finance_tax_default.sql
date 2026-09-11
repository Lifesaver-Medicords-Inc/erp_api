-- ============================================================
-- Default every Business Partner's Finance tax to
--     TAX CODE = VAT      TAX = 12
-- wherever it is blank.
--
-- WHY: management decision (2026-09-11). VAT is the company-wide
-- standard - spec 4.5.3 lists VAT at 12 as "standard & default" - and
-- the partners imported from QERP arrived with no tax code at all.
-- The QERP import (05 load_bpi_from_qerp.sql) now writes the same
-- default, so this is for records that already exist.
--
-- 12, not 0.12 and not '12%': spec 4.5.3 stores a rate as a
-- percentage number, and every consumer divides by 100.
--
-- WHAT IS TOUCHED - finance rows of branches that exist:
--   1. tax code blank AND tax blank  -> VAT / 12
--   2. tax code VAT   AND tax blank  -> tax 12
--   3. a branch with no finance row  -> a new row, VAT / 12
-- Anything someone entered is left exactly as it is (S1, NON-VAT, a
-- rate already filled in...). Finance rows whose branch no longer
-- exists are ignored.
--
-- REVERSIBLE: every touched row is recorded in
-- zz_bpi_finance_tax_backup with its previous values. Rows with
-- action = 'inserted' did not exist before and can simply be deleted
-- by finance_id.
--
-- IDEMPOTENT: a second run finds nothing blank and changes 0 rows.
-- SQL Server 2012 compatible.
-- ============================================================
SET NOCOUNT ON;
SET XACT_ABORT ON;

IF OBJECT_ID(N'zz_bpi_finance_tax_backup', N'U') IS NULL
    CREATE TABLE zz_bpi_finance_tax_backup (
        finance_id    BIGINT        NOT NULL,
        action        NVARCHAR(10)  NOT NULL,   -- 'updated' | 'inserted'
        old_tax_code  NVARCHAR(MAX) NULL,
        old_tax       NVARCHAR(MAX) NULL,
        backed_up_at  DATETIME      NOT NULL DEFAULT GETDATE()
    );
GO

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @both_blank INT, @vat_no_rate INT, @added INT;
DECLARE @new TABLE (finance_id BIGINT);

BEGIN TRAN;

-- 1. no tax code and no rate
INSERT INTO zz_bpi_finance_tax_backup (finance_id, action, old_tax_code, old_tax)
SELECT f.finance_id, N'updated', f.finance_tax_code, f.finance_tax
FROM tbl_bpi_finance f
JOIN tbl_bpi_general g ON g.id = f.finance_branch_id
WHERE LTRIM(RTRIM(ISNULL(f.finance_tax_code, N''))) = N''
  AND LTRIM(RTRIM(ISNULL(f.finance_tax, N''))) = N'';

UPDATE f
SET finance_tax_code = N'VAT', finance_tax = N'12'
FROM tbl_bpi_finance f
JOIN tbl_bpi_general g ON g.id = f.finance_branch_id
WHERE LTRIM(RTRIM(ISNULL(f.finance_tax_code, N''))) = N''
  AND LTRIM(RTRIM(ISNULL(f.finance_tax, N''))) = N'';
SET @both_blank = @@ROWCOUNT;

-- 2. VAT with no rate
INSERT INTO zz_bpi_finance_tax_backup (finance_id, action, old_tax_code, old_tax)
SELECT f.finance_id, N'updated', f.finance_tax_code, f.finance_tax
FROM tbl_bpi_finance f
JOIN tbl_bpi_general g ON g.id = f.finance_branch_id
WHERE UPPER(LTRIM(RTRIM(ISNULL(f.finance_tax_code, N'')))) = N'VAT'
  AND LTRIM(RTRIM(ISNULL(f.finance_tax, N''))) = N'';

UPDATE f
SET finance_tax = N'12'
FROM tbl_bpi_finance f
JOIN tbl_bpi_general g ON g.id = f.finance_branch_id
WHERE UPPER(LTRIM(RTRIM(ISNULL(f.finance_tax_code, N'')))) = N'VAT'
  AND LTRIM(RTRIM(ISNULL(f.finance_tax, N''))) = N'';
SET @vat_no_rate = @@ROWCOUNT;

-- 3. branches that have no finance row at all
INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_tax_code, finance_tax)
OUTPUT inserted.finance_id INTO @new (finance_id)
SELECT g.based_id, g.id, N'VAT', N'12'
FROM tbl_bpi_general g
WHERE NOT EXISTS (SELECT 1 FROM tbl_bpi_finance f WHERE f.finance_branch_id = g.id);
SET @added = @@ROWCOUNT;

INSERT INTO zz_bpi_finance_tax_backup (finance_id, action)
SELECT finance_id, N'inserted' FROM @new;

COMMIT;

PRINT N'blank -> VAT / 12          : ' + CAST(@both_blank AS NVARCHAR(20));
PRINT N'VAT with no rate -> 12     : ' + CAST(@vat_no_rate AS NVARCHAR(20));
PRINT N'finance rows added         : ' + CAST(@added AS NVARCHAR(20));
GO
