-- ============================================================
-- Give every Business Partner that has no industry the industry
-- OTH = Others, at BOTH levels:
--     tbl_bpi_industries         company level  (the header INDUSTRIES)
--     tbl_bpi_branch_industries  branch level   (BRANCH INDUSTRY, spec 4.1.3)
--
-- WHY: vw_get_bpi_list ends with
--     WHERE ISNULL(c.industry_names, '') <> ''
-- so a company with no industry is dropped from the BPI form's list
-- entirely - it cannot even be opened to add one. The QERP import had
-- no industry column, and 18 partners in the rehearsal data were
-- already invisible for the same reason. Management's decision: such
-- partners get OTH = Others rather than the view being changed.
--
-- NOTE: OTH is NOT one of the 26 industries listed in spec 17.3. It is
-- a deliberate management addition; the spec should gain it.
--
-- Both levels, because spec 4.1.1 defines the header INDUSTRIES as the
-- aggregate of all branches and 4.1.3 makes BRANCH INDUSTRY required.
--
-- IDEMPOTENT: only rows with no industry at all are touched, and OTH
-- itself is only inserted if its code is missing. Never overrides an
-- industry someone chose.
-- ============================================================
SET NOCOUNT ON;

IF NOT EXISTS (SELECT 1 FROM tbl_setup_bpi_industries WHERE code = N'OTH')
    INSERT INTO tbl_setup_bpi_industries (code, name) VALUES (N'OTH', N'Others');

DECLARE @oth BIGINT;
SELECT TOP 1 @oth = id FROM tbl_setup_bpi_industries WHERE code = N'OTH' ORDER BY id;

INSERT INTO tbl_bpi_industries (bpi_id, industry_id)
SELECT b.id, @oth
FROM tbl_bpi b
WHERE NOT EXISTS (SELECT 1 FROM tbl_bpi_industries i WHERE i.bpi_id = b.id);
DECLARE @companies INT = @@ROWCOUNT;

INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id)
SELECT g.id, @oth
FROM tbl_bpi_general g
WHERE NOT EXISTS (SELECT 1 FROM tbl_bpi_branch_industries i WHERE i.bpi_general_id = g.id);
DECLARE @branches INT = @@ROWCOUNT;

PRINT N'OTH industry id        : ' + CAST(@oth AS NVARCHAR(20));
PRINT N'companies given OTH    : ' + CAST(@companies AS NVARCHAR(20));
PRINT N'branches given OTH     : ' + CAST(@branches AS NVARCHAR(20));
GO
