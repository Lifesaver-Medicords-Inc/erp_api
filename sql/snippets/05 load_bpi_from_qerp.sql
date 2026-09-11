-- ============================================================
-- Lightspeed ERP - Business Partner Info load
-- Generated from 'data migration excel format/BPI data sheet (from QERP)_1.xlsx'
-- Sheets used: 'Entity Clean data 32326', 'BPI HEADER (PARENTS)',
--              'BPI HEADER (CHILDREN)'   (the two Entity UNEDITED/Clean
--              sheets are deliberately ignored).
-- Target: SQL Server 2012.   Parents: 17   Children: 14
--
-- HOW TO RUN
--     sqlcmd -S <server> -U sa -P <pw> -d <database> -b -I -i <this file>
--     Run load_setup_reference_data.sql FIRST - entity types and payment
--     terms are resolved by CODE and must already exist.
--
-- STRUCTURE. A 'parent' becomes the company, a 'child' becomes another
-- branch under that same company:
--     tbl_bpi           one row per PARENT      - the company
--     tbl_bpi_general   one row per parent AND per child - the branches;
--                       the parent's own row carries is_main = 1
--     tbl_bpi_entity    CUS / SUP link per branch
--     tbl_bpi_address   the branch address
--     tbl_bpi_finance   payment terms and tax code per branch
--
-- WHERE THE CUSTOMER/SUPPLIER FLAGS COME FROM - and why not from the
-- header sheets: in BOTH 'BPI HEADER' sheets the Is Customer, Is Supplier
-- and Active columns are filled with the row's PHONE NUMBER (e.g. ALPHALAND
-- reads Is Customer = '337-2031'), and one row reads #N/A. They are a fill
-- error, not data. Every flag here is therefore read from
-- 'Entity Clean data 32326', which holds proper TRUE/FALSE.
--
-- customer_code / supplier_code are generated the way the API generates
-- them (generateCustomerCode / generateSupplierCode in bpi_entity_service
-- .go): C#0001 / S#0001, continuing from the highest already issued. QERP's
-- own entity codes are NOT reused as customer codes - they are kept in
-- tbl_bpi_general.notes so the QERP record remains traceable.
--
-- INDUSTRY: the workbook has none, and vw_get_bpi_list HIDES any company
-- without one. Every imported company and branch is given OTH = Others
-- (management decision; OTH is not among spec 17.3's 26 values). OTH is
-- inserted into tbl_setup_bpi_industries first if its code is missing.
--
-- TAX: a branch with no QERP tax code gets TAX CODE = VAT and TAX = 12
-- (management decision 2026-09-11; spec 4.5.3 stores a rate as a
-- percentage number - 12, not 0.12 or '12%'). A QERP tax code is kept.
--
-- transaction_type (TRADE / NON TRADE) is left unset: the workbook has no
-- column for it and inventing one per partner would be a guess.
--
-- IDEMPOTENT by company name: a partner whose name already exists is left
-- untouched, so re-running adds only what is new.
--
-- SKIPPED:
--   EASTWEST CREDIT CARD             blank Entity Code in the CHILDREN sheet
--
-- PAYMENT TERM MAPPINGS APPLIED:
--   30 DAYS    -> NET30
--   COD        -> COD
--   PDC 45     -> PDC45
-- ============================================================
SET NOCOUNT ON;
GO

DECLARE @bpi BIGINT, @gen BIGINT, @cust INT, @sup INT, @oth BIGINT;
IF NOT EXISTS (SELECT 1 FROM tbl_setup_bpi_industries WHERE code = N'OTH')
    INSERT INTO tbl_setup_bpi_industries (code, name) VALUES (N'OTH', N'Others');
SELECT TOP 1 @oth = id FROM tbl_setup_bpi_industries WHERE code = N'OTH' ORDER BY id;
SELECT @cust = ISNULL(MAX(TRY_CAST(SUBSTRING(customer_code, 3, LEN(customer_code)) AS INT)), 0)
  FROM tbl_bpi_general WHERE customer_code LIKE 'C#%';
SELECT @sup  = ISNULL(MAX(TRY_CAST(SUBSTRING(supplier_code, 3, LEN(supplier_code)) AS INT)), 0)
  FROM tbl_bpi_general WHERE supplier_code LIKE 'S#%';

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'CONSOLIDATED PRIME DEV. CORP.')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'CONSOLIDATED PRIME DEV. CORP.', N'005-649-159-001', NULL, NULL, N'OFFICE');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'CONSOLIDATED PRIME DEV. CORP.', N'OFFICE', 1, N'LOCAL',
        NULL, NULL, NULL, 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: 0498');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'SM CITY DASMARIÑAS GOVERNOR''S DRIVE BRGY. SAMPALOK, DASMA, CAVITE', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'A. LUMERAN PLUMBING')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'A. LUMERAN PLUMBING', N'113-338-823-000', N'456-1789', NULL, N'OFFICE');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'A. LUMERAN PLUMBING', N'OFFICE', 1, N'LOCAL',
        N'456-1789', NULL, N'456-1789', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0001');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'24C OFELIA VILLAGE ROAD PROJ 8 Q.C.', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'ACE PACKAGING CO.')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'ACE PACKAGING CO.', N'000-372-636-000', N'983-2780', NULL, N'OFFICE');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ACE PACKAGING CO.', N'OFFICE', 1, N'LOCAL',
        N'983-2780', NULL, N'983-2945', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0002');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'73 RD 7 GSIS HILLS  TALIPAPA NOVALICHES', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'ACE TOWER CONDOMINIUM')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'ACE TOWER CONDOMINIUM', N'201-035-022-000', NULL, NULL, N'OFFICE');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ACE TOWER CONDOMINIUM', N'OFFICE', 1, N'LOCAL',
        NULL, NULL, NULL, 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0003');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'211 BANAWE ST., QUEZON CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'ALPHALAND CORPORATION')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'ALPHALAND CORPORATION', N'001-746-612-000', N'337-2031', NULL, N'J. CESTONA');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ALPHALAND CORPORATION', N'J. CESTONA', 1, N'LOCAL',
        N'337-2031', NULL, N'310-5329', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0006');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'2258 CHINO ROCES AVE., MAKATI CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
    -- child: 007-9ALPHALAND BALESIN ISLAND CLUB
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'007-9ALPHALAND BALESIN ISLAND CLUB', N'J. CESTONA', 0, N'LOCAL',
        NULL, NULL, NULL, 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: ABIC000001');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'Real Quezon', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
    -- child: ALPHALAND BALESIN ISLAND CLUB INC.
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ALPHALAND BALESIN ISLAND CLUB INC.', N'J. CESTONA', 0, N'LOCAL',
        N'337-2031', NULL, N'310-5329', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0004');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'2258 CHINO ROCES AVE., MAKATI CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
    -- child: ALPHALAND BALESIN ISLAND RESORT CORP.
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ALPHALAND BALESIN ISLAND RESORT CORP.', N'J. CESTONA', 0, N'LOCAL',
        N'337-2031', NULL, N'310-5329', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0005');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'2258 CHINO ROCES AVE., MAKATI CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
    -- child: ALPHALAND DEVELOPMENT INC.
    SET @cust = @cust + 1;
    SET @sup = @sup + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ALPHALAND DEVELOPMENT INC.', N'J. CESTONA', 0, N'LOCAL',
        N'337-2031', NULL, N'310-5329', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), 'S#' + RIGHT('0000' + CAST(@sup AS NVARCHAR(20)), 4), N'QERP entity code: C0007');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'SUP';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'2258 CHINO ROCES AVE., MAKATI CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
    -- child: ALPHALAND MAKATI PLACE INC
    SET @cust = @cust + 1;
    SET @sup = @sup + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ALPHALAND MAKATI PLACE INC', N'J. CESTONA', 0, N'LOCAL',
        N'337-2031', NULL, N'310-5329', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), 'S#' + RIGHT('0000' + CAST(@sup AS NVARCHAR(20)), 4), N'QERP entity code: C0008');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'SUP';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'AYALA AVE COR MALUGAY ST MAKATI', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
    -- child: ALPHALAND UKIYO INC.
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ALPHALAND UKIYO INC.', N'J. CESTONA', 0, N'LOCAL',
        N'337-2031', NULL, N'310-5329', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0009');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'2258 CHINO ROCES AVE., MAKATI CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'ALSOFI ENGINEERING WORK')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'ALSOFI ENGINEERING WORK', N'104-002-703-000', N'453-2459', NULL, N'OFFICE');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ALSOFI ENGINEERING WORK', N'OFFICE', 1, N'LOCAL',
        N'453-2459', NULL, NULL, 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0010');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'53 TANDANG SORA ST., QUEZON CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'AMERICAN WATER TECHNOLOGIES INC.')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'AMERICAN WATER TECHNOLOGIES INC.', N'004-761-096-000', NULL, NULL, N'JUANITO SY');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AMERICAN WATER TECHNOLOGIES INC.', N'JUANITO SY', 1, N'LOCAL',
        NULL, NULL, NULL, 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0013');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'660C ML QUEZON ST CASUPINGAN MANDAUE', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'APO REFRIGERATION AND AIRCONDITIONING')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'APO REFRIGERATION AND AIRCONDITIONING', N'167-314-985-000', N'821-8170', NULL, N'OFFICE');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'APO REFRIGERATION AND AIRCONDITIONING', N'OFFICE', 1, N'LOCAL',
        N'821-8170', NULL, NULL, 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0014');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'13 MICHAEL RUA ST BETTER LIVING PNQUE', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'ARMSTRONG PLUMBING CORP.')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'ARMSTRONG PLUMBING CORP.', N'002-623-799-000', N'732-5991', NULL, N'L. SALANGA');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ARMSTRONG PLUMBING CORP.', N'L. SALANGA', 1, N'LOCAL',
        N'732-5991', NULL, N'732-60-25', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0015');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'1679 MALABON ST.,STA CRUZ, MANILA', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'ARSEY INDUSTRIAL CENTER, INC.')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'ARSEY INDUSTRIAL CENTER, INC.', N'008-582-257-001', N'733-1811', NULL, N'OFFICE');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ARSEY INDUSTRIAL CENTER, INC.', N'OFFICE', 1, N'LOCAL',
        N'733-1811', NULL, N'733-2143', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0016');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'627 T. ALONZO ST., STA CRUZ, MANILA', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'ASCOTT MAKATI')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'ASCOTT MAKATI', N'006-616-580-000', NULL, NULL, N'L. SALANGA');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ASCOTT MAKATI', N'L. SALANGA', 1, N'LOCAL',
        NULL, NULL, NULL, 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0017');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'B3 GLORIETTA 4 AYALA CENTER MAKATI CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'ASIAN CARMAKERS CORPORATION')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'ASIAN CARMAKERS CORPORATION', N'000-695-510-000', N'897-0445', NULL, N'J. CESTONA');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ASIAN CARMAKERS CORPORATION', N'J. CESTONA', 1, N'LOCAL',
        N'897-0445', NULL, N'897-0445', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0018');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'LOT 5 KALAYAAN AVE, MAKATI CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'ASIAN DYNASTY IMPORT-EXPORT CORP.')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'ASIAN DYNASTY IMPORT-EXPORT CORP.', N'204-195-434-000', N'829-6484', NULL, N'JUANITO SY');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ASIAN DYNASTY IMPORT-EXPORT CORP.', N'JUANITO SY', 1, N'LOCAL',
        N'829-6484', NULL, N'804-2554', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0019');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'2677 CAPINPIN ST., BANGKAL MAKATI  CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'AVESCO MARKETING CORPORATION')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'AVESCO MARKETING CORPORATION', N'000-400-152-001', N'912-8881', NULL, N'OFFICE');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AVESCO MARKETING CORPORATION', N'OFFICE', 1, N'LOCAL',
        N'912-8881', NULL, N'912-8881', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0020');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'AURORA BLVD YALE CUBAO QUEZON CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'AVISLEY HARDWARE CORPORATION')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'AVISLEY HARDWARE CORPORATION', N'007-371-532-000', N'361-6794', NULL, N'JUANITO SY');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AVISLEY HARDWARE CORPORATION', N'JUANITO SY', 1, N'LOCAL',
        N'361-6794', NULL, N'365-6799', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0025');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'372 A. BONIFACIO AVE., BALINTAWAK QUEZON', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'AYALA LAND INC.')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'AYALA LAND INC.', N'000-153-790-000', N'818-7471', NULL, N'J. CESTONA');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AYALA LAND INC.', N'J. CESTONA', 1, N'LOCAL',
        N'818-7471', NULL, N'818-0149', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0026');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'TRIANGLE TOWER ONE & EXCHANGE PLAZA, AYALA, MAKATI B1 ADMIN OFFICE GLORIETTA 5, AYALA CENTER, MAKATI CITY TRIANGLE TOWER ONE  & EXCHANGE PLAZA, AYALA, MAKATI', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
    -- child: ALVEO LAND CORPORATION
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'ALVEO LAND CORPORATION', N'L. SALANGA', 0, N'LOCAL',
        N'798-8565', NULL, NULL, 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0011');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'2ND FLR. B7 & B8 BONI HIGH ST., TAGUIG', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
    -- child: AMAIA LAND CORP.
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AMAIA LAND CORP.', N'L. SALANGA', 0, N'LOCAL',
        NULL, NULL, NULL, 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0012');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'10F AYALA LIFE FGU CTR, MUNTINLUPA CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
    -- child: AVIDA LAND CORP.
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AVIDA LAND CORP.', N'L. SALANGA', 0, N'LOCAL',
        N'988-4999', NULL, N'840-4543', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0021');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'909 40TH ST. NORTH, BONIFACIO TRIANGLE 909 40TH ST. NORTH, BONIFACIO TRIANGLE, BONIFACIO GLOBAL CITY, TAGUIG', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
    -- child: AVIDA TOWER NEW MANILA CONDO CORP.
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AVIDA TOWER NEW MANILA CONDO CORP.', N'J. CESTONA', 0, N'LOCAL',
        N'621-6176', NULL, N'621-6176', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0022');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'4F AYALA AVE., MAKATI CITY', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
    -- child: AVIDA TOWER SAN LAZARO CONDO CORP.
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AVIDA TOWER SAN LAZARO CONDO CORP.', N'L. SALANGA', 0, N'LOCAL',
        N'753-1181', NULL, N'753-1181', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0023');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'FELIX HUERTAS ST., STA CRUZ, MANILA', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
    -- child: AVIDA TOWER SUCAT CONDO. CORP.
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AVIDA TOWER SUCAT CONDO. CORP.', N'L. SALANGA', 0, N'LOCAL',
        N'799-4402', NULL, NULL, 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0024');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'DR.A SANTOS AVE. BRGY SAN DIONISIO PNQUE', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
    -- child: AYALA LIFE-FGU CENTER ALABANG CONDO CORP
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AYALA LIFE-FGU CENTER ALABANG CONDO CORP', N'L. SALANGA', 0, N'LOCAL',
        N'809-6084', NULL, NULL, 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0027');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'5TH FLR GLORIETTA 4 AYALA CENTER MAKATI', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'COD'), N'VAT', N'12');
    -- child: AYALA PROPERTY MANAGEMENT CORPORATION
    SET @cust = @cust + 1;
    SET @sup = @sup + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AYALA PROPERTY MANAGEMENT CORPORATION', N'J. CESTONA', 0, N'LOCAL',
        N'818-7471', NULL, N'818-0149', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), 'S#' + RIGHT('0000' + CAST(@sup AS NVARCHAR(20)), 4), N'QERP entity code: C0028');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'SUP';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'5TH FLR GLORIETTA 4 AYALA CENTER MAKATI', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'NET30'), N'VAT', N'12');
END

IF NOT EXISTS (SELECT 1 FROM tbl_bpi WHERE name = N'AYLAN CONSTRUCTION & TRADING')
BEGIN
    INSERT INTO tbl_bpi (name, tin, main_tel_no, main_website, sales_id)
    VALUES (N'AYLAN CONSTRUCTION & TRADING', N'103-979-294-000', N'740-0181', NULL, N'OFFICE');
    SET @bpi = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_industries (bpi_id, industry_id) VALUES (@bpi, @oth);
    SET @cust = @cust + 1;
    INSERT INTO tbl_bpi_general (based_id, branch_name, sales_id, is_main, class_name,
        branch_tel_no, branch_website, fax_no, customer_code, supplier_code, notes)
    VALUES (@bpi, N'AYLAN CONSTRUCTION & TRADING', N'OFFICE', 1, N'LOCAL',
        N'740-0181', NULL, N'740-0181', 'C#' + RIGHT('0000' + CAST(@cust AS NVARCHAR(20)), 4), NULL, N'QERP entity code: C0029');
    SET @gen = SCOPE_IDENTITY();
    INSERT INTO tbl_bpi_branch_industries (bpi_general_id, industry_id) VALUES (@gen, @oth);
    INSERT INTO tbl_bpi_entity (bpi_general_id, entity_id)
    SELECT @gen, id FROM tbl_setup_bpi_entity WHERE code = N'CUS';
    INSERT INTO tbl_bpi_address (based_id, branch_id, location, is_deleted)
    VALUES (@bpi, @gen, N'163 AMORANTO ST., LALOMA QUEZON CITY 163 N.S. AMORANTO ST., LA LOMA Q.C.', 0);
    INSERT INTO tbl_bpi_finance (finance_based_id, finance_branch_id, finance_payment_terms_id, finance_tax_code, finance_tax)
    VALUES (@bpi, @gen, (SELECT TOP 1 id FROM tbl_setup_payment_terms WHERE code = N'PDC45'), N'VAT', N'12');
END

PRINT N'BPI load complete';
GO

SELECT 'tbl_bpi' AS [table], COUNT(*) AS rows_now FROM tbl_bpi;
SELECT 'tbl_bpi_general', COUNT(*) FROM tbl_bpi_general;
SELECT 'tbl_bpi_entity', COUNT(*) FROM tbl_bpi_entity;
SELECT 'tbl_bpi_address', COUNT(*) FROM tbl_bpi_address;
SELECT 'tbl_bpi_finance', COUNT(*) FROM tbl_bpi_finance;
SELECT 'branches with no entity type' AS check_, COUNT(*) AS n FROM tbl_bpi_general g
  WHERE NOT EXISTS (SELECT 1 FROM tbl_bpi_entity e WHERE e.bpi_general_id = g.id);
GO
