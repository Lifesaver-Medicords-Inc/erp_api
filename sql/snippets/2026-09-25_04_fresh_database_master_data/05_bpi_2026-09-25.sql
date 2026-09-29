-- STEP 05 of 05 - BPI - business partners
-- Run the numbered scripts in this folder in order, each on its own: 01, 02 ... 05.
-- Run after:  04_boms_and_templates_2026-09-25.sql
-- Run before: restarting the API (see AFTER THE LAST FILE below)
-- Every step is a TRIAL RUN until you set @commit = 1; it must end by printing "committed".

-- Master data for a NEW database: BPI - business partners
-- ============================================================================
-- Every customer and supplier: branches, addresses, contacts, industries, entity, finance
-- terms, supplier items, accreditation records and history. Accreditation FILES are not in
-- the database - copy the API's files folder across separately.
--
-- Generated 2026-09-25 15:42 from DESKTOP-CS0MRJ5 / Lightspeed_test_fresh by sql/snippets/generate_fresh_database_master_data.sql.
-- Regenerate rather than edit: the data below is a snapshot of that database on that day.
--
-- WHAT THE BATCH IS FOR
--   A database the API has just created has its tables but no master data. The five files
--   of this batch give it the source's master data with the SAME ids the source uses, so every
--   reference between them (item -> brand, BOM line -> item, BPI -> payment terms...) still lines
--   up and scripts written against the source work here too. NOT carried: transactions (quotes,
--   orders, stock, accounting documents), the z_*_at audit history, and what the API seeds for
--   itself (access modules, position access, calendar cost types).
--
-- WHAT IT REFUSES TO DO
--   It loads into EMPTY tables only. The rows the API seeds into a new database - the DEFAULT
--   item name, class and brand, the UNIT placeholder, the DEFAULT-PLACEHOLDER item, the starter
--   template, and the Dispatcher and AR/AP Cashier positions with their access - are removed and
--   replaced. Anything else already in a table stops the script before it writes a thing, so it
--   cannot damage a working database, and running a committed file again changes nothing.
--
-- WHEN THE SCHEMA HAS MOVED ON
--   Columns are matched by NAME. A new nullable column in the database is simply left empty.
--   It stops, listing what and why, for: a table the database does not have (set
--   @skip_missing_tables = 1 to load without it); a column holding data here that the database
--   does not have (@allow_dropped_columns = 1 to load without it); a NOT NULL column with no
--   default that this file cannot fill; text longer than its column; a fraction headed for a
--   whole-number column. After loading it re-checks every reference and rolls everything back
--   if more rows point at nothing than did in the source.
--
-- AFTER THE LAST FILE
--   Restart the API. On start it re-creates the Dispatcher and AR/AP Cashier positions,
--   gives every position its starting access, and the item and BPI lists are read fresh.
--
-- Tables in this file, in load order (rows):
--   tbl_bpi (20)
--   tbl_bpi_general (36)
--   tbl_bpi_accreditation (1)
--   tbl_bpi_address (37)
--   tbl_bpi_branch_industries (43)
--   tbl_bpi_contacts (6)
--   tbl_bpi_entity (40)
--   tbl_bpi_finance (51)
--   tbl_bpi_history (35)
--   tbl_bpi_industries (26)
--   tbl_bpi_items (2)
--
-- HOW TO RUN: open it in SSMS on the NEW database and execute, or
--   sqlcmd -S <server> -d <database> -b -i <this file>
-- With @commit = 0 everything is checked, loaded and counted, then rolled back.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;                 -- 1 = keep the changes, 0 = trial run
DECLARE @skip_missing_tables bit = 0;    -- see WHEN THE SCHEMA HAS MOVED ON
DECLARE @allow_dropped_columns bit = 0;  -- see WHEN THE SCHEMA HAS MOVED ON

IF OBJECT_ID('tempdb..#plan') IS NOT NULL DROP TABLE #plan;
IF OBJECT_ID('tempdb..#refs') IS NOT NULL DROP TABLE #refs;
IF OBJECT_ID('tempdb..#m_tbl_bpi') IS NOT NULL DROP TABLE #m_tbl_bpi;
IF OBJECT_ID('tempdb..#m_tbl_bpi_general') IS NOT NULL DROP TABLE #m_tbl_bpi_general;
IF OBJECT_ID('tempdb..#m_tbl_bpi_accreditation') IS NOT NULL DROP TABLE #m_tbl_bpi_accreditation;
IF OBJECT_ID('tempdb..#m_tbl_bpi_address') IS NOT NULL DROP TABLE #m_tbl_bpi_address;
IF OBJECT_ID('tempdb..#m_tbl_bpi_branch_industries') IS NOT NULL DROP TABLE #m_tbl_bpi_branch_industries;
IF OBJECT_ID('tempdb..#m_tbl_bpi_contacts') IS NOT NULL DROP TABLE #m_tbl_bpi_contacts;
IF OBJECT_ID('tempdb..#m_tbl_bpi_entity') IS NOT NULL DROP TABLE #m_tbl_bpi_entity;
IF OBJECT_ID('tempdb..#m_tbl_bpi_finance') IS NOT NULL DROP TABLE #m_tbl_bpi_finance;
IF OBJECT_ID('tempdb..#m_tbl_bpi_history') IS NOT NULL DROP TABLE #m_tbl_bpi_history;
IF OBJECT_ID('tempdb..#m_tbl_bpi_industries') IS NOT NULL DROP TABLE #m_tbl_bpi_industries;
IF OBJECT_ID('tempdb..#m_tbl_bpi_items') IS NOT NULL DROP TABLE #m_tbl_bpi_items;

-- What goes where. placeholder = the API's own rows, the only ones this may delete.
CREATE TABLE #plan (ord int NOT NULL PRIMARY KEY, target sysname NOT NULL, stage sysname NULL,
                    rows_in_file int NOT NULL, placeholder nvarchar(max) NULL);
INSERT INTO #plan (ord, target, stage, rows_in_file, placeholder) VALUES
    (10, N'tbl_bpi', N'#m_tbl_bpi', 20, NULL),
    (20, N'tbl_bpi_general', N'#m_tbl_bpi_general', 36, NULL),
    (1001, N'tbl_bpi_accreditation', N'#m_tbl_bpi_accreditation', 1, NULL),
    (1002, N'tbl_bpi_address', N'#m_tbl_bpi_address', 37, NULL),
    (1003, N'tbl_bpi_branch_industries', N'#m_tbl_bpi_branch_industries', 43, NULL),
    (1004, N'tbl_bpi_contacts', N'#m_tbl_bpi_contacts', 6, NULL),
    (1005, N'tbl_bpi_entity', N'#m_tbl_bpi_entity', 40, NULL),
    (1006, N'tbl_bpi_finance', N'#m_tbl_bpi_finance', 51, NULL),
    (1007, N'tbl_bpi_history', N'#m_tbl_bpi_history', 35, NULL),
    (1008, N'tbl_bpi_industries', N'#m_tbl_bpi_industries', 26, NULL),
    (1009, N'tbl_bpi_items', N'#m_tbl_bpi_items', 2, NULL);

-- References re-checked after loading, with how many already dangled in the source.
CREATE TABLE #refs (child sysname NOT NULL, col sysname NOT NULL, parent sysname NOT NULL, pcol sysname NOT NULL,
                    dangling_in_source int NOT NULL, dangling_here int NULL);
INSERT INTO #refs (child, col, parent, pcol, dangling_in_source) VALUES
    (N'tbl_bpi_accreditation', N'based_id', N'tbl_bpi', N'id', 0),
    (N'tbl_bpi_accreditation', N'branch_id', N'tbl_bpi_general', N'id', 0),
    (N'tbl_bpi_address', N'based_id', N'tbl_bpi', N'id', 0),
    (N'tbl_bpi_address', N'branch_id', N'tbl_bpi_general', N'id', 0),
    (N'tbl_bpi_branch_industries', N'bpi_general_id', N'tbl_bpi_general', N'id', 0),
    (N'tbl_bpi_branch_industries', N'industry_id', N'tbl_setup_bpi_industries', N'id', 0),
    (N'tbl_bpi_contacts', N'based_id', N'tbl_bpi', N'id', 0),
    (N'tbl_bpi_contacts', N'branch_id', N'tbl_bpi_general', N'id', 0),
    (N'tbl_bpi_contacts', N'position', N'tbl_position', N'id', 0),
    (N'tbl_bpi_entity', N'bpi_general_id', N'tbl_bpi_general', N'id', 0),
    (N'tbl_bpi_entity', N'entity_id', N'tbl_setup_bpi_entity', N'id', 0),
    (N'tbl_bpi_finance', N'finance_account_id', N'tbl_setup_chart_of_accounts', N'id', 1),
    (N'tbl_bpi_finance', N'finance_based_id', N'tbl_bpi', N'id', 0),
    (N'tbl_bpi_finance', N'finance_branch_id', N'tbl_bpi_general', N'id', 0),
    (N'tbl_bpi_finance', N'finance_payment_terms_id', N'tbl_setup_payment_terms', N'id', 0),
    (N'tbl_bpi_general', N'based_id', N'tbl_bpi', N'id', 0),
    (N'tbl_bpi_general', N'social', N'tbl_setup_social', N'id', 0),
    (N'tbl_bpi_history', N'based_id', N'tbl_bpi', N'id', 9),
    (N'tbl_bpi_industries', N'bpi_id', N'tbl_bpi', N'id', 0),
    (N'tbl_bpi_industries', N'industry_id', N'tbl_setup_bpi_industries', N'id', 0),
    (N'tbl_bpi_items', N'based_id', N'tbl_bpi', N'id', 0),
    (N'tbl_bpi_items', N'branch_id', N'tbl_bpi_general', N'id', 0),
    (N'tbl_bpi_items', N'item_account_id', N'tbl_setup_chart_of_accounts', N'id', 0),
    (N'tbl_bpi_items', N'item_id', N'tbl_setup_item', N'id', 0),
    (N'tbl_bpi_items', N'payment_terms_id', N'tbl_setup_payment_terms', N'id', 0);

-- tbl_bpi (20 rows) -----------------------------------------------------
CREATE TABLE #m_tbl_bpi ([id] bigint NULL, [sales_id] nvarchar(max) NULL, [name] nvarchar(max) NULL, [main_website] nvarchar(max) NULL, [tin] nvarchar(max) NULL, [main_tel_no] nvarchar(max) NULL);
INSERT INTO #m_tbl_bpi ([id], [sales_id], [name], [main_website], [tin], [main_tel_no]) VALUES
    (1, N'OFFICE', N'CONSOLIDATED PRIME DEV. CORP.', NULL, N'005-649-159-001', NULL),
    (2, N'OFFICE', N'A. LUMERAN PLUMBING', NULL, N'113-338-823-000', N'456-1789'),
    (3, N'OFFICE', N'ACE PACKAGING CO.', NULL, N'000-372-636-000', N'983-2780'),
    (4, N'OFFICE', N'ACE TOWER CONDOMINIUM', NULL, N'201-035-022-000', NULL),
    (5, N'Julie CESTONA', N'ALPHALAND CORPORATION', NULL, N'001-746-612-000', N'337-2031'),
    (6, N'OFFICE', N'ALSOFI ENGINEERING WORK', NULL, N'104-002-703-000', N'453-2459'),
    (7, N'JUANITO SY', N'AMERICAN WATER TECHNOLOGIES INC.', NULL, N'004-761-096-000', NULL),
    (8, N'OFFICE', N'APO REFRIGERATION AND AIRCONDITIONING', NULL, N'167-314-985-000', N'821-8170'),
    (9, N'L. SALANGA', N'ARMSTRONG PLUMBING CORP.', NULL, N'002-623-799-000', N'732-5991'),
    (10, N'OFFICE', N'ARSEY INDUSTRIAL CENTER, INC.', NULL, N'008-582-257-001', N'733-1811'),
    (11, N'L. SALANGA', N'ASCOTT MAKATI', NULL, N'006-616-580-000', NULL),
    (12, N'Julie CESTONA', N'ASIAN CARMAKERS CORPORATION', NULL, N'000-695-510-000', N'897-0445'),
    (13, N'JUANITO SY', N'ASIAN DYNASTY IMPORT-EXPORT CORP.', NULL, N'204-195-434-000', N'829-6484'),
    (14, N'OFFICE', N'AVESCO MARKETING CORPORATION', NULL, N'000-400-152-001', N'912-8881'),
    (15, N'JUANITO SY', N'AVISLEY HARDWARE CORPORATION', NULL, N'007-371-532-000', N'361-6794'),
    (16, N'Julie CESTONA', N'AYALA LAND INC.', NULL, N'000-153-790-000', N'818-7471'),
    (17, N'OFFICE', N'AYLAN CONSTRUCTION & TRADING', NULL, N'103-979-294-000', N'740-0181'),
    (18, N'Cai', N'TEST CUSTOMER 1', N'testc1.com', N'009-122-026-00000', N'0912-345-6789'),
    (24, N'Cai', N'TEST SUPPLIER 1', N'tests1.com', N'123-456-789-00000', N'0912-345-6789'),
    (31, N'Rachelle Robles', N'TEST CUSTOMER 2', N'testcus2.com', N'009-122-123-00000', N'0998-765-4321');

-- tbl_bpi_general (36 rows) ---------------------------------------------
CREATE TABLE #m_tbl_bpi_general ([id] bigint NULL, [based_id] bigint NULL, [branch_name] nvarchar(100) NULL, [sales_id] nvarchar(max) NULL, [is_main] bit NULL, [social] bigint NULL, [transaction_type] nvarchar(max) NULL, [class_name] nvarchar(max) NULL, [branch_tel_no] nvarchar(max) NULL, [branch_website] nvarchar(max) NULL, [customer_code] nvarchar(max) NULL, [supplier_code] nvarchar(max) NULL, [fax_no] nvarchar(max) NULL, [notes] nvarchar(max) NULL);
INSERT INTO #m_tbl_bpi_general ([id], [based_id], [branch_name], [sales_id], [is_main], [social], [transaction_type], [class_name], [branch_tel_no], [branch_website], [customer_code], [supplier_code], [fax_no], [notes]) VALUES
    (1, 1, N'CONSOLIDATED PRIME DEV. CORP.', N'OFFICE', 1, NULL, NULL, N'LOCAL', NULL, NULL, N'C#0001', NULL, NULL, N'QERP entity code: 0498'),
    (2, 2, N'A. LUMERAN PLUMBING', N'OFFICE', 1, NULL, NULL, N'LOCAL', N'456-1789', NULL, N'C#0002', NULL, N'456-1789', N'QERP entity code: C0001'),
    (3, 3, N'ACE PACKAGING CO.', N'OFFICE', 1, NULL, NULL, N'LOCAL', N'983-2780', NULL, N'C#0003', NULL, N'983-2945', N'QERP entity code: C0002'),
    (4, 4, N'ACE TOWER CONDOMINIUM', N'OFFICE', 1, NULL, NULL, N'LOCAL', NULL, NULL, N'C#0004', NULL, NULL, N'QERP entity code: C0003'),
    (5, 5, N'ALPHALAND CORPORATION', N'Julie CESTONA', 1, NULL, NULL, N'LOCAL', N'337-2031', NULL, N'C#0005', NULL, N'310-5329', N'QERP entity code: C0006'),
    (6, 5, N'007-9ALPHALAND BALESIN ISLAND CLUB', N'Julie CESTONA', 0, NULL, NULL, N'LOCAL', NULL, NULL, N'C#0006', NULL, NULL, N'QERP entity code: ABIC000001'),
    (7, 5, N'ALPHALAND BALESIN ISLAND CLUB INC.', N'Julie CESTONA', 0, NULL, NULL, N'LOCAL', N'337-2031', NULL, N'C#0007', NULL, N'310-5329', N'QERP entity code: C0004'),
    (8, 5, N'ALPHALAND BALESIN ISLAND RESORT CORP.', N'Julie CESTONA', 0, NULL, NULL, N'LOCAL', N'337-2031', NULL, N'C#0008', NULL, N'310-5329', N'QERP entity code: C0005'),
    (9, 5, N'ALPHALAND DEVELOPMENT INC.', N'Julie CESTONA', 0, NULL, NULL, N'LOCAL', N'337-2031', NULL, N'C#0009', N'S#0001', N'310-5329', N'QERP entity code: C0007'),
    (10, 5, N'ALPHALAND MAKATI PLACE INC', N'Julie CESTONA', 0, NULL, NULL, N'LOCAL', N'337-2031', NULL, N'C#0010', N'S#0002', N'310-5329', N'QERP entity code: C0008'),
    (11, 5, N'ALPHALAND UKIYO INC.', N'Julie CESTONA', 0, NULL, NULL, N'LOCAL', N'337-2031', NULL, N'C#0011', NULL, N'310-5329', N'QERP entity code: C0009'),
    (12, 6, N'ALSOFI ENGINEERING WORK', N'OFFICE', 1, NULL, NULL, N'LOCAL', N'453-2459', NULL, N'C#0012', NULL, NULL, N'QERP entity code: C0010'),
    (13, 7, N'AMERICAN WATER TECHNOLOGIES INC.', N'JUANITO SY', 1, NULL, NULL, N'LOCAL', NULL, NULL, N'C#0013', NULL, NULL, N'QERP entity code: C0013'),
    (14, 8, N'APO REFRIGERATION AND AIRCONDITIONING', N'OFFICE', 1, NULL, NULL, N'LOCAL', N'821-8170', NULL, N'C#0014', NULL, NULL, N'QERP entity code: C0014'),
    (15, 9, N'ARMSTRONG PLUMBING CORP.', N'L. SALANGA', 1, NULL, NULL, N'LOCAL', N'732-5991', NULL, N'C#0015', NULL, N'732-60-25', N'QERP entity code: C0015'),
    (16, 10, N'ARSEY INDUSTRIAL CENTER, INC.', N'OFFICE', 1, NULL, NULL, N'LOCAL', N'733-1811', NULL, N'C#0016', NULL, N'733-2143', N'QERP entity code: C0016'),
    (17, 11, N'ASCOTT MAKATI', N'L. SALANGA', 1, NULL, NULL, N'LOCAL', NULL, NULL, N'C#0017', NULL, NULL, N'QERP entity code: C0017'),
    (18, 12, N'ASIAN CARMAKERS CORPORATION', N'Julie CESTONA', 1, NULL, NULL, N'LOCAL', N'897-0445', NULL, N'C#0018', NULL, N'897-0445', N'QERP entity code: C0018'),
    (19, 13, N'ASIAN DYNASTY IMPORT-EXPORT CORP.', N'JUANITO SY', 1, NULL, NULL, N'LOCAL', N'829-6484', NULL, N'C#0019', NULL, N'804-2554', N'QERP entity code: C0019'),
    (20, 14, N'AVESCO MARKETING CORPORATION', N'OFFICE', 1, NULL, NULL, N'LOCAL', N'912-8881', NULL, N'C#0020', NULL, N'912-8881', N'QERP entity code: C0020'),
    (21, 15, N'AVISLEY HARDWARE CORPORATION', N'JUANITO SY', 1, NULL, NULL, N'LOCAL', N'361-6794', NULL, N'C#0021', NULL, N'365-6799', N'QERP entity code: C0025'),
    (22, 16, N'AYALA LAND INC.', N'Julie CESTONA', 1, NULL, NULL, N'LOCAL', N'818-7471', NULL, N'C#0022', NULL, N'818-0149', N'QERP entity code: C0026'),
    (23, 16, N'ALVEO LAND CORPORATION', N'L. SALANGA', 0, NULL, NULL, N'LOCAL', N'798-8565', NULL, N'C#0023', NULL, NULL, N'QERP entity code: C0011'),
    (24, 16, N'AMAIA LAND CORP.', N'L. SALANGA', 0, NULL, NULL, N'LOCAL', NULL, NULL, N'C#0024', NULL, NULL, N'QERP entity code: C0012'),
    (25, 16, N'AVIDA LAND CORP.', N'L. SALANGA', 0, NULL, NULL, N'LOCAL', N'988-4999', NULL, N'C#0025', NULL, N'840-4543', N'QERP entity code: C0021'),
    (26, 16, N'AVIDA TOWER NEW MANILA CONDO CORP.', N'Julie CESTONA', 0, NULL, NULL, N'LOCAL', N'621-6176', NULL, N'C#0026', NULL, N'621-6176', N'QERP entity code: C0022'),
    (27, 16, N'AVIDA TOWER SAN LAZARO CONDO CORP.', N'L. SALANGA', 0, NULL, NULL, N'LOCAL', N'753-1181', NULL, N'C#0027', NULL, N'753-1181', N'QERP entity code: C0023'),
    (28, 16, N'AVIDA TOWER SUCAT CONDO. CORP.', N'L. SALANGA', 0, NULL, NULL, N'LOCAL', N'799-4402', NULL, N'C#0028', NULL, NULL, N'QERP entity code: C0024'),
    (29, 16, N'AYALA LIFE-FGU CENTER ALABANG CONDO CORP', N'L. SALANGA', 0, NULL, NULL, N'LOCAL', N'809-6084', NULL, N'C#0029', NULL, NULL, N'QERP entity code: C0027'),
    (30, 16, N'AYALA PROPERTY MANAGEMENT CORPORATION', N'Julie CESTONA', 0, NULL, NULL, N'LOCAL', N'818-7471', NULL, N'C#0030', N'S#0003', N'818-0149', N'QERP entity code: C0028'),
    (31, 17, N'AYLAN CONSTRUCTION & TRADING', N'OFFICE', 1, NULL, NULL, N'LOCAL', N'740-0181', NULL, N'C#0031', NULL, N'740-0181', N'QERP entity code: C0029'),
    (32, 18, N'TEST CUSTOMER 1', N'Cai', 0, 8, N'TRADE', N'LOCAL', N'0912-345-6789', N'testc1.com', N'C#0032', N'', N'', N''),
    (38, 24, N'TEST SUPPLIER 1', N'Cai', 1, 0, N'TRADE', N'LOCAL', N'0912-345-6789', N'tests1.com', N'', N'S#0004', N'', N''),
    (45, 18, N'TEST SUPPLIER 2', N'Cai', 0, 0, N'TRADE', N'FOREIGN', N'', N'', N'', N'S#0005', N'', N''),
    (46, 24, N'TEST CUSSUP 1', N'Rachelle Robles', 0, 2, N'TRADE', N'LOCAL', N'', N'', N'C#0033', N'S#0006', N'', N''),
    (47, 31, N'TEST CUSTOMER 2', N'Rachelle Robles', 0, 0, N'TRADE', N'LOCAL', N'0998-765-4321', N'testcus2.com', N'C#0034', N'', N'', N'');

-- tbl_bpi_accreditation (1 rows) ---------------------------------------
CREATE TABLE #m_tbl_bpi_accreditation ([id] bigint NULL, [based_id] bigint NULL, [branch_id] bigint NULL, [date_added] nvarchar(max) NULL, [file_path] nvarchar(max) NULL, [file_name] nvarchar(max) NULL, [accreditation_added_by] nvarchar(max) NULL);
INSERT INTO #m_tbl_bpi_accreditation ([id], [based_id], [branch_id], [date_added], [file_path], [file_name], [accreditation_added_by]) VALUES
    (1, 18, 32, N'12/09/2026 3:48:35 pm', N'./files/1789199451028516500.docx', N'SM ILOILO Master VFD1 to VFD2.docx', N'C. ');

-- tbl_bpi_address (37 rows) ---------------------------------------------
CREATE TABLE #m_tbl_bpi_address ([id] bigint NULL, [based_id] bigint NULL, [branch_id] bigint NULL, [location] nvarchar(max) NULL, [is_deleted] bit NULL);
INSERT INTO #m_tbl_bpi_address ([id], [based_id], [branch_id], [location], [is_deleted]) VALUES
    (1, 1, 1, N'SM CITY DASMARIÑAS GOVERNOR''S DRIVE BRGY. SAMPALOK, DASMA, CAVITE', 0),
    (2, 2, 2, N'24C OFELIA VILLAGE ROAD PROJ 8 Q.C.', 0),
    (3, 3, 3, N'73 RD 7 GSIS HILLS  TALIPAPA NOVALICHES', 0),
    (4, 4, 4, N'211 BANAWE ST., QUEZON CITY', 0),
    (5, 5, 5, N'2258 CHINO ROCES AVE., MAKATI CITY', 0),
    (6, 5, 6, N'Real Quezon', 0),
    (7, 5, 7, N'2258 CHINO ROCES AVE., MAKATI CITY', 0),
    (8, 5, 8, N'2258 CHINO ROCES AVE., MAKATI CITY', 0),
    (9, 5, 9, N'2258 CHINO ROCES AVE., MAKATI CITY', 0),
    (10, 5, 10, N'AYALA AVE COR MALUGAY ST MAKATI', 0),
    (11, 5, 11, N'2258 CHINO ROCES AVE., MAKATI CITY', 0),
    (12, 6, 12, N'53 TANDANG SORA ST., QUEZON CITY', 0),
    (13, 7, 13, N'660C ML QUEZON ST CASUPINGAN MANDAUE', 0),
    (14, 8, 14, N'13 MICHAEL RUA ST BETTER LIVING PNQUE', 0),
    (15, 9, 15, N'1679 MALABON ST.,STA CRUZ, MANILA', 0),
    (16, 10, 16, N'627 T. ALONZO ST., STA CRUZ, MANILA', 0),
    (17, 11, 17, N'B3 GLORIETTA 4 AYALA CENTER MAKATI CITY', 0),
    (18, 12, 18, N'LOT 5 KALAYAAN AVE, MAKATI CITY', 0),
    (19, 13, 19, N'2677 CAPINPIN ST., BANGKAL MAKATI  CITY', 0),
    (20, 14, 20, N'AURORA BLVD YALE CUBAO QUEZON CITY', 0),
    (21, 15, 21, N'372 A. BONIFACIO AVE., BALINTAWAK QUEZON', 0),
    (22, 16, 22, N'TRIANGLE TOWER ONE & EXCHANGE PLAZA, AYALA, MAKATI B1 ADMIN OFFICE GLORIETTA 5, AYALA CENTER, MAKATI CITY TRIANGLE TOWER ONE  & EXCHANGE PLAZA, AYALA, MAKATI', 0),
    (23, 16, 23, N'2ND FLR. B7 & B8 BONI HIGH ST., TAGUIG', 0),
    (24, 16, 24, N'10F AYALA LIFE FGU CTR, MUNTINLUPA CITY', 0),
    (25, 16, 25, N'909 40TH ST. NORTH, BONIFACIO TRIANGLE 909 40TH ST. NORTH, BONIFACIO TRIANGLE, BONIFACIO GLOBAL CITY, TAGUIG', 0),
    (26, 16, 26, N'4F AYALA AVE., MAKATI CITY', 0),
    (27, 16, 27, N'FELIX HUERTAS ST., STA CRUZ, MANILA', 0),
    (28, 16, 28, N'DR.A SANTOS AVE. BRGY SAN DIONISIO PNQUE', 0),
    (29, 16, 29, N'5TH FLR GLORIETTA 4 AYALA CENTER MAKATI', 0),
    (30, 16, 30, N'5TH FLR GLORIETTA 4 AYALA CENTER MAKATI', 0),
    (31, 17, 31, N'163 AMORANTO ST., LALOMA QUEZON CITY 163 N.S. AMORANTO ST., LA LOMA Q.C.', 0),
    (32, 18, 32, N'test cus 1 add st.', 0),
    (33, 24, 38, N'test st2', 0),
    (34, 18, 45, N'test sup 2 add st.', 0),
    (35, 24, 46, N'test cussup 1 add st.', 0),
    (36, 24, 46, N'test cussup 2 add st.', 0),
    (37, 31, 47, N'test cus 2 add st.', 0);

-- tbl_bpi_branch_industries (43 rows) -----------------------------------
CREATE TABLE #m_tbl_bpi_branch_industries ([id] bigint NULL, [bpi_general_id] bigint NULL, [industry_id] bigint NULL);
INSERT INTO #m_tbl_bpi_branch_industries ([id], [bpi_general_id], [industry_id]) VALUES
    (1, 1, 41),
    (2, 2, 41),
    (3, 3, 41),
    (4, 4, 41),
    (5, 5, 41),
    (6, 6, 41),
    (7, 7, 41),
    (8, 8, 41),
    (9, 9, 41),
    (10, 10, 41),
    (11, 11, 41),
    (12, 12, 41),
    (13, 13, 41),
    (14, 14, 41),
    (15, 15, 41),
    (16, 16, 41),
    (17, 17, 41),
    (18, 18, 41),
    (19, 19, 41),
    (20, 20, 41),
    (21, 21, 41),
    (22, 22, 41),
    (23, 23, 41),
    (24, 24, 41),
    (25, 25, 41),
    (26, 26, 41),
    (27, 27, 41),
    (28, 28, 41),
    (29, 29, 41),
    (30, 30, 41),
    (32, 31, 41),
    (49, 45, 18),
    (78, 38, 1),
    (79, 38, 3),
    (80, 38, 4),
    (81, 38, 6),
    (82, 38, 43),
    (83, 46, 24),
    (85, 47, 15),
    (86, 47, 17),
    (94, 32, 1),
    (95, 32, 3),
    (96, 32, 4);

-- tbl_bpi_contacts (6 rows) --------------------------------------------
CREATE TABLE #m_tbl_bpi_contacts ([id] bigint NULL, [based_id] bigint NULL, [branch_id] bigint NULL, [number] nvarchar(max) NULL, [name] nvarchar(max) NULL, [email] nvarchar(max) NULL, [preferences] nvarchar(max) NULL, [position] bigint NULL, [is_default_contact] bit NULL, [contact_notes] nvarchar(max) NULL);
INSERT INTO #m_tbl_bpi_contacts ([id], [based_id], [branch_id], [number], [name], [email], [preferences], [position], [is_default_contact], [contact_notes]) VALUES
    (1, 18, 32, N'0912-345-6789', N'testconcus 1', N'', N'VBR', 1, 1, N''),
    (2, 24, 38, N'0912-345-6789', N'testconsup1', N'', N'EML', 3, 1, N''),
    (3, 18, 45, N'0912-345-6789', N'testconsup 2', N'', N'VBR', 3, 1, N''),
    (4, 24, 46, N'', N'testconcussup 1', N'cussup1@gmail.com', N'EML', 1, 0, N''),
    (5, 24, 46, N'0912-345-6789', N'testconcussup 2', N'', N'VBR', 20002, 0, N''),
    (6, 31, 47, N'0998-765-4321', N'testconcus 2', N'', N'WSP', 20002, 1, N'');

-- tbl_bpi_entity (40 rows) ----------------------------------------------
CREATE TABLE #m_tbl_bpi_entity ([id] bigint NULL, [bpi_general_id] bigint NULL, [entity_id] bigint NULL);
INSERT INTO #m_tbl_bpi_entity ([id], [bpi_general_id], [entity_id]) VALUES
    (1, 1, 2),
    (2, 2, 2),
    (3, 3, 2),
    (4, 4, 2),
    (5, 5, 2),
    (6, 6, 2),
    (7, 7, 2),
    (8, 8, 2),
    (9, 9, 2),
    (10, 9, 1),
    (11, 10, 2),
    (12, 10, 1),
    (13, 11, 2),
    (14, 12, 2),
    (15, 13, 2),
    (16, 14, 2),
    (17, 15, 2),
    (18, 16, 2),
    (19, 17, 2),
    (20, 18, 2),
    (21, 19, 2),
    (22, 20, 2),
    (23, 21, 2),
    (24, 22, 2),
    (25, 23, 2),
    (26, 24, 2),
    (27, 25, 2),
    (28, 26, 2),
    (29, 27, 2),
    (30, 28, 2),
    (31, 29, 2),
    (32, 30, 2),
    (33, 30, 1),
    (35, 31, 2),
    (52, 45, 1),
    (71, 38, 1),
    (72, 46, 1),
    (73, 46, 2),
    (75, 47, 2),
    (79, 32, 2);

-- tbl_bpi_finance (51 rows) ---------------------------------------------
CREATE TABLE #m_tbl_bpi_finance ([finance_id] bigint NULL, [finance_based_id] bigint NULL, [finance_account_id] bigint NULL, [finance_payment_terms_id] bigint NULL, [finance_branch_id] bigint NULL, [finance_tax_code] nvarchar(max) NULL, [finance_tax] nvarchar(max) NULL);
INSERT INTO #m_tbl_bpi_finance ([finance_id], [finance_based_id], [finance_account_id], [finance_payment_terms_id], [finance_branch_id], [finance_tax_code], [finance_tax]) VALUES
    (1, 1, NULL, 32, 1, N'VAT', N'12'),
    (2, 2, NULL, 29, 2, N'VAT', N'12'),
    (3, 3, NULL, 29, 3, N'VAT', N'12'),
    (4, 4, NULL, 29, 4, N'VAT', N'12'),
    (5, 5, NULL, 32, 5, N'VAT', N'12'),
    (6, 5, NULL, 32, 6, N'VAT', N'12'),
    (7, 5, NULL, 32, 7, N'VAT', N'12'),
    (8, 5, NULL, 32, 8, N'VAT', N'12'),
    (9, 5, NULL, 32, 9, N'VAT', N'12'),
    (10, 5, NULL, 32, 10, N'VAT', N'12'),
    (11, 5, NULL, 32, 11, N'VAT', N'12'),
    (12, 6, NULL, 29, 12, N'VAT', N'12'),
    (13, 7, NULL, 32, 13, N'VAT', N'12'),
    (14, 8, NULL, 29, 14, N'VAT', N'12'),
    (15, 9, NULL, 29, 15, N'VAT', N'12'),
    (16, 10, NULL, 29, 16, N'VAT', N'12'),
    (17, 11, NULL, 32, 17, N'VAT', N'12'),
    (18, 12, NULL, 32, 18, N'VAT', N'12'),
    (19, 13, NULL, 29, 19, N'VAT', N'12'),
    (20, 14, NULL, 29, 20, N'VAT', N'12'),
    (21, 15, NULL, 29, 21, N'VAT', N'12'),
    (22, 16, NULL, 32, 22, N'VAT', N'12'),
    (23, 16, NULL, 29, 23, N'VAT', N'12'),
    (24, 16, NULL, 32, 24, N'VAT', N'12'),
    (25, 16, NULL, 32, 25, N'VAT', N'12'),
    (26, 16, NULL, 32, 26, N'VAT', N'12'),
    (27, 16, NULL, 32, 27, N'VAT', N'12'),
    (28, 16, NULL, 29, 28, N'VAT', N'12'),
    (29, 16, NULL, 29, 29, N'VAT', N'12'),
    (30, 16, NULL, 32, 30, N'VAT', N'12'),
    (31, 17, NULL, 35, 31, N'VAT', N'12'),
    (32, 18, 4, 28, 32, N'', N''),
    (33, 24, 0, 0, 38, N'', N''),
    (34, 24, 0, 0, 0, N'', N''),
    (35, 24, 0, 0, 0, N'', N''),
    (36, 24, 0, 0, 0, N'', N''),
    (37, 24, 0, 0, 0, N'', N''),
    (38, 24, 0, 0, 0, N'', N''),
    (39, 24, 0, 0, 0, N'', N''),
    (40, 24, 0, 0, 0, N'', N''),
    (41, 24, 0, 0, 0, N'', N''),
    (42, 24, 0, 0, 0, N'', N''),
    (43, 24, 0, 0, 0, N'', N''),
    (44, 24, 0, 0, 0, N'', N''),
    (45, 24, 0, 0, 0, N'', N''),
    (46, 24, 0, 0, 0, N'', N''),
    (47, 24, 0, 0, 0, N'', N''),
    (48, 24, 0, 0, 0, N'', N''),
    (49, 24, 0, 0, 0, N'', N''),
    (50, 24, 0, 0, 0, N'', N''),
    (51, 31, 0, 0, 47, N'', N'');

-- tbl_bpi_history (35 rows) ---------------------------------------------
CREATE TABLE #m_tbl_bpi_history ([id] bigint NULL, [based_id] bigint NULL, [date] nvarchar(max) NULL, [actions] nvarchar(max) NULL, [child_type] nvarchar(max) NULL, [edit_by] nvarchar(max) NULL);
INSERT INTO #m_tbl_bpi_history ([id], [based_id], [date], [actions], [child_type], [edit_by]) VALUES
    (1, 31, N'2026-09-11 13:32:45', N'update', N'General', N'Gids'),
    (2, 32, N'2026-09-12 15:50:50', N'create', N'General', N'Cai'),
    (3, 18, N'2026-09-12 15:50:50', N'create', N'Contacts', N'Cai'),
    (4, 18, N'2026-09-12 15:50:51', N'create', N'Address', N'Cai'),
    (5, 18, N'2026-09-12 15:50:51', N'create', N'Finance', N'Cai'),
    (6, 18, N'2026-09-12 15:50:51', N'create', N'Accreditation', N'Cai'),
    (7, 38, N'2026-09-12 16:10:35', N'create', N'General', N'Cai'),
    (8, 24, N'2026-09-12 16:10:35', N'create', N'Contacts', N'Cai'),
    (9, 24, N'2026-09-12 16:10:35', N'create', N'Address', N'Cai'),
    (10, 24, N'2026-09-12 16:10:35', N'create', N'Finance', N'Cai'),
    (16, 32, N'2026-09-12 16:28:35', N'update', N'Contacts', N'Cai'),
    (17, 24, N'2026-09-12 17:04:01', N'create', N'Items', N'Cai'),
    (18, 24, N'2026-09-12 17:04:01', N'create', N'Items', N'Cai'),
    (19, 24, N'2026-09-12 17:04:01', N'create', N'Items', N'Cai'),
    (20, 45, N'2026-09-14 16:50:35', N'create', N'General', N'Cai'),
    (21, 32, N'2026-09-14 16:52:22', N'update', N'Address', N'Cai'),
    (22, 46, N'2026-09-15 11:02:24', N'create', N'General', N'Rachelle Robles'),
    (23, 24, N'2026-09-15 11:02:24', N'delete', N'Items', N'Rachelle Robles'),
    (24, 46, N'2026-09-15 11:03:14', N'update', N'General', N'Rachelle Robles'),
    (25, 46, N'2026-09-15 11:50:04', N'update', N'General', N'Rachelle Robles'),
    (26, 24, N'2026-09-15 11:50:04', N'create', N'Items', N'Rachelle Robles'),
    (27, 24, N'2026-09-15 11:50:04', N'create', N'Items', N'Rachelle Robles'),
    (28, 24, N'2026-09-15 16:34:49', N'delete', N'Items', N'Gids'),
    (29, 24, N'2026-09-15 16:34:49', N'delete', N'Items', N'Gids'),
    (30, 24, N'2026-09-15 16:34:49', N'create', N'Items', N'Gids'),
    (31, 24, N'2026-09-17 08:56:58', N'delete', N'Items', N'Cai'),
    (32, 24, N'2026-09-17 13:56:49', N'create', N'Items', N'Cai'),
    (33, 24, N'2026-09-17 15:09:22', N'delete', N'Items', N'Cai'),
    (34, 24, N'2026-09-17 15:09:22', N'create', N'Items', N'Cai'),
    (35, 24, N'2026-09-17 15:09:22', N'create', N'Items', N'Cai'),
    (36, 18, N'2026-09-19 11:44:48', N'update', N'Finance', N'Cai'),
    (37, 47, N'2026-09-23 11:24:47', N'create', N'General', N'Rachelle Robles'),
    (38, 31, N'2026-09-23 11:24:47', N'create', N'Contacts', N'Rachelle Robles'),
    (39, 31, N'2026-09-23 11:24:47', N'create', N'Address', N'Rachelle Robles'),
    (40, 31, N'2026-09-23 11:24:47', N'create', N'Finance', N'Rachelle Robles');

-- tbl_bpi_industries (26 rows) ------------------------------------------
CREATE TABLE #m_tbl_bpi_industries ([id] bigint NULL, [bpi_id] bigint NULL, [industry_id] bigint NULL);
INSERT INTO #m_tbl_bpi_industries ([id], [bpi_id], [industry_id]) VALUES
    (1, 1, 41),
    (2, 2, 41),
    (3, 3, 41),
    (4, 4, 41),
    (5, 5, 41),
    (6, 6, 41),
    (7, 7, 41),
    (8, 8, 41),
    (9, 9, 41),
    (10, 10, 41),
    (11, 11, 41),
    (12, 12, 41),
    (13, 13, 41),
    (14, 14, 41),
    (15, 15, 41),
    (16, 16, 41),
    (18, 17, 41),
    (80, 24, 1),
    (81, 24, 3),
    (82, 24, 6),
    (83, 24, 43),
    (85, 31, 15),
    (86, 31, 17),
    (94, 18, 1),
    (95, 18, 3),
    (96, 18, 4);

-- tbl_bpi_items (2 rows) -----------------------------------------------
CREATE TABLE #m_tbl_bpi_items ([id] bigint NULL, [based_id] bigint NULL, [branch_id] bigint NULL, [payment_terms_id] bigint NULL, [item_account_id] bigint NULL, [item_id] bigint NULL, [canvass_id] bigint NULL, [tax_code] nvarchar(max) NULL, [item_tax_code] nvarchar(max) NULL, [price] float NULL, [notes] nvarchar(max) NULL, [is_deleted] bit NULL);
INSERT INTO #m_tbl_bpi_items ([id], [based_id], [branch_id], [payment_terms_id], [item_account_id], [item_id], [canvass_id], [tax_code], [item_tax_code], [price], [notes], [is_deleted]) VALUES
    (8, 24, 46, 23, 0, 1251, 0, N'VAT', N'', 0, N'', 0),
    (9, 24, 46, 23, 0, 1256, 0, N'VAT', N'', 0, N'', 0);

-- ============================================================================
-- THE LOADER - the same in every file of this batch; only the data above differs.
-- It checks everything first, and writes nothing unless every check passes.
-- ============================================================================
IF OBJECT_ID('tempdb..#problems') IS NOT NULL DROP TABLE #problems;
IF OBJECT_ID('tempdb..#notes')    IS NOT NULL DROP TABLE #notes;
IF OBJECT_ID('tempdb..#cols')     IS NOT NULL DROP TABLE #cols;
IF OBJECT_ID('tempdb..#loaded')   IS NOT NULL DROP TABLE #loaded;
CREATE TABLE #problems (target sysname NOT NULL, problem nvarchar(2000) NOT NULL);
CREATE TABLE #notes    (target sysname NOT NULL, note nvarchar(2000) NOT NULL);
CREATE TABLE #cols     (ord int NOT NULL, col sysname NOT NULL, colid int NOT NULL, load_it bit NOT NULL);
CREATE TABLE #loaded   (ord int NOT NULL PRIMARY KEY, target sysname NOT NULL, rows_in_file int NOT NULL,
                        placeholders_removed int NOT NULL, rows_loaded int NOT NULL);

DECLARE @ord int, @target sysname, @stage sysname, @rows_in_file int, @placeholder nvarchar(max),
        @tobj int, @sobj int, @sql nvarchar(max), @n int, @m int, @col sysname,
        @ttype sysname, @tlen int, @stype sysname, @collist nvarchar(max), @identity bit,
        @child sysname, @parent sysname, @pcol sysname;

-- 1. Check ----------------------------------------------------------------------
DECLARE plan_cur CURSOR LOCAL FAST_FORWARD FOR
    SELECT ord, target, stage, rows_in_file, placeholder FROM #plan ORDER BY ord;
OPEN plan_cur;
FETCH NEXT FROM plan_cur INTO @ord, @target, @stage, @rows_in_file, @placeholder;
WHILE @@FETCH_STATUS = 0
BEGIN
    SET @tobj = OBJECT_ID(N'dbo.' + QUOTENAME(@target), N'U');
    SET @sobj = OBJECT_ID(N'tempdb..' + @stage);

    -- Every row of this file has to have reached its staging table. A run that stopped
    -- part-way (SSMS carries on after an error) would otherwise load a short table.
    IF @stage IS NOT NULL
    BEGIN
        SET @n = NULL;
        IF @sobj IS NOT NULL
        BEGIN
            SET @sql = N'SELECT @n = COUNT(*) FROM ' + @stage + N';';
            EXEC sp_executesql @sql, N'@n int OUTPUT', @n = @n OUTPUT;
        END
        IF ISNULL(@n, -1) <> @rows_in_file
            INSERT INTO #problems VALUES (@target, N'only ' + ISNULL(CAST(@n AS nvarchar(20)), N'0') + N' of the '
                + CAST(@rows_in_file AS nvarchar(20)) + N' rows in this file were read - run the whole file again, from the top');
    END

    IF @tobj IS NULL
    BEGIN
        IF @stage IS NULL
            INSERT INTO #notes VALUES (@target, N'not in this database - nothing to clear');
        ELSE IF @skip_missing_tables = 1
            INSERT INTO #notes VALUES (@target, N'NOT LOADED - the table is not in this database ('
                + CAST(@rows_in_file AS nvarchar(20)) + N' rows left out because @skip_missing_tables = 1)');
        ELSE
            INSERT INTO #problems VALUES (@target, N'the table is not in this database. Start the API against it once so it '
                + N'builds its tables. If the table was retired on purpose, set @skip_missing_tables = 1 to load without it.');
    END
    ELSE
    BEGIN
        -- Anything in the table except the API's own placeholder rows means this is not a new database.
        SET @sql = N'SELECT @n = COUNT(*) FROM dbo.' + QUOTENAME(@target) + N'; SELECT @m = COUNT(*) FROM dbo.'
                 + QUOTENAME(@target) + N' WHERE ' + ISNULL(@placeholder, N'1 = 0') + N';';
        EXEC sp_executesql @sql, N'@n int OUTPUT, @m int OUTPUT', @n = @n OUTPUT, @m = @m OUTPUT;
        IF @n > @m
            INSERT INTO #problems VALUES (@target, CAST(@n - @m AS nvarchar(20)) + N' row(s) already in the table. This loads a NEW '
                + N'database only - it never adds to or overwrites data that is there. If this file was already committed '
                + N'here, there is nothing left to do.');

        IF @sobj IS NOT NULL
        BEGIN
            INSERT INTO #cols (ord, col, colid, load_it)
            SELECT @ord, s.name, s.column_id,
                   CASE WHEN EXISTS (SELECT 1 FROM sys.columns t
                                      WHERE t.object_id = @tobj AND t.is_computed = 0
                                        AND t.name = s.name COLLATE DATABASE_DEFAULT) THEN 1 ELSE 0 END
              FROM tempdb.sys.columns s
             WHERE s.object_id = @sobj;

            -- A column this file carries that the database no longer has.
            DECLARE col_cur CURSOR LOCAL FAST_FORWARD FOR
                SELECT col FROM #cols WHERE ord = @ord AND load_it = 0 ORDER BY colid;
            OPEN col_cur;
            FETCH NEXT FROM col_cur INTO @col;
            WHILE @@FETCH_STATUS = 0
            BEGIN
                SET @sql = N'SELECT @n = COUNT(*) FROM ' + @stage + N' WHERE ' + QUOTENAME(@col) + N' IS NOT NULL;';
                EXEC sp_executesql @sql, N'@n int OUTPUT', @n = @n OUTPUT;
                IF @n > 0 AND @allow_dropped_columns = 0
                    INSERT INTO #problems VALUES (@target, N'column ' + @col + N' holds ' + CAST(@n AS nvarchar(20))
                        + N' value(s) but this database has no such column. Set @allow_dropped_columns = 1 to load without it.');
                ELSE
                    INSERT INTO #notes VALUES (@target, N'column ' + @col + N' left out - this database has no such column ('
                        + CAST(@n AS nvarchar(20)) + N' value(s) not loaded)');
                FETCH NEXT FROM col_cur INTO @col;
            END
            CLOSE col_cur;
            DEALLOCATE col_cur;

            -- A column the database insists on that this file does not carry.
            INSERT INTO #problems
            SELECT @target, N'column ' + t.name + N' must be filled (NOT NULL, no default) but this file has no such column'
              FROM sys.columns t
             WHERE t.object_id = @tobj AND t.is_nullable = 0 AND t.is_identity = 0 AND t.is_computed = 0
               AND t.default_object_id = 0
               AND NOT EXISTS (SELECT 1 FROM #cols c
                                WHERE c.ord = @ord AND c.load_it = 1 AND c.col = t.name COLLATE DATABASE_DEFAULT);

            -- Values that would not survive the trip: text longer than the column, or a
            -- fraction headed for a whole-number column. Both would be cut, not refused.
            DECLARE fit_cur CURSOR LOCAL FAST_FORWARD FOR
                SELECT c.col, tt.name, t.max_length, st.name
                  FROM #cols c
                  JOIN sys.columns t  ON t.object_id = @tobj AND t.name = c.col COLLATE DATABASE_DEFAULT
                  JOIN sys.types   tt ON tt.user_type_id = t.user_type_id
                  JOIN tempdb.sys.columns s ON s.object_id = @sobj AND s.column_id = c.colid
                  JOIN tempdb.sys.types  st ON st.user_type_id = s.user_type_id
                 WHERE c.ord = @ord AND c.load_it = 1;
            OPEN fit_cur;
            FETCH NEXT FROM fit_cur INTO @col, @ttype, @tlen, @stype;
            WHILE @@FETCH_STATUS = 0
            BEGIN
                SET @sql = NULL;
                IF @ttype IN (N'nvarchar', N'nchar') AND @tlen <> -1
                    SET @tlen = @tlen / 2;
                IF @ttype IN (N'nvarchar', N'nchar', N'varchar', N'char') AND @tlen <> -1
                    SET @sql = N'SELECT @n = COUNT(*) FROM ' + @stage + N' WHERE LEN(' + QUOTENAME(@col) + N' + N''.'') - 1 > '
                             + CAST(@tlen AS nvarchar(20)) + N';';
                ELSE IF @ttype IN (N'bigint', N'int', N'smallint', N'tinyint') AND @stype IN (N'float', N'real', N'decimal', N'numeric')
                    SET @sql = N'SELECT @n = COUNT(*) FROM ' + @stage + N' WHERE ' + QUOTENAME(@col) + N' <> FLOOR(' + QUOTENAME(@col) + N');';
                IF @sql IS NOT NULL
                BEGIN
                    EXEC sp_executesql @sql, N'@n int OUTPUT', @n = @n OUTPUT;
                    IF @n > 0
                        INSERT INTO #problems VALUES (@target, N'column ' + @col + N': ' + CAST(@n AS nvarchar(20))
                            + CASE WHEN @ttype IN (N'bigint', N'int', N'smallint', N'tinyint')
                                   THEN N' value(s) have a fraction, and this database keeps whole numbers only (' + @ttype + N')'
                                   ELSE N' value(s) are longer than this database allows (' + @ttype + N'('
                                        + CAST(@tlen AS nvarchar(20)) + N'))' END);
                END
                FETCH NEXT FROM fit_cur INTO @col, @ttype, @tlen, @stype;
            END
            CLOSE fit_cur;
            DEALLOCATE fit_cur;
        END
    END
    FETCH NEXT FROM plan_cur INTO @ord, @target, @stage, @rows_in_file, @placeholder;
END
CLOSE plan_cur;
DEALLOCATE plan_cur;

IF EXISTS (SELECT 1 FROM #problems)
BEGIN
    SELECT N'NOTHING WAS WRITTEN' AS result, target AS [table], problem FROM #problems ORDER BY target;
    IF EXISTS (SELECT 1 FROM #notes)
        SELECT target AS [table], note FROM #notes ORDER BY target;
    RAISERROR(N'This database cannot take this file (see the Results tab). Nothing was written.', 16, 1);
    RETURN;
END

-- 2. Write ----------------------------------------------------------------------
BEGIN TRANSACTION;

INSERT INTO #loaded (ord, target, rows_in_file, placeholders_removed, rows_loaded)
SELECT ord, target, rows_in_file, 0, 0 FROM #plan;

-- The API's placeholder rows go first, children before parents.
DECLARE ph_cur CURSOR LOCAL FAST_FORWARD FOR
    SELECT ord, target, placeholder FROM #plan
     WHERE placeholder IS NOT NULL AND OBJECT_ID(N'dbo.' + QUOTENAME(target), N'U') IS NOT NULL
     ORDER BY ord DESC;
OPEN ph_cur;
FETCH NEXT FROM ph_cur INTO @ord, @target, @placeholder;
WHILE @@FETCH_STATUS = 0
BEGIN
    SET @sql = N'DELETE FROM dbo.' + QUOTENAME(@target) + N' WHERE ' + @placeholder + N'; SET @n = @@ROWCOUNT;';
    EXEC sp_executesql @sql, N'@n int OUTPUT', @n = @n OUTPUT;
    UPDATE #loaded SET placeholders_removed = @n WHERE ord = @ord;
    FETCH NEXT FROM ph_cur INTO @ord, @target, @placeholder;
END
CLOSE ph_cur;
DEALLOCATE ph_cur;

-- Then the rows, parents before children, each keeping the id it has in the source.
DECLARE load_cur CURSOR LOCAL FAST_FORWARD FOR
    SELECT ord, target, stage FROM #plan
     WHERE stage IS NOT NULL AND OBJECT_ID(N'dbo.' + QUOTENAME(target), N'U') IS NOT NULL
     ORDER BY ord;
OPEN load_cur;
FETCH NEXT FROM load_cur INTO @ord, @target, @stage;
WHILE @@FETCH_STATUS = 0
BEGIN
    SET @collist = STUFF((SELECT N', ' + QUOTENAME(col) FROM #cols
                           WHERE ord = @ord AND load_it = 1 ORDER BY colid FOR XML PATH('')), 1, 2, N'');
    SET @identity = CASE WHEN EXISTS (
            SELECT 1 FROM sys.columns t
              JOIN #cols c ON c.ord = @ord AND c.load_it = 1 AND c.col = t.name COLLATE DATABASE_DEFAULT
             WHERE t.object_id = OBJECT_ID(N'dbo.' + QUOTENAME(@target)) AND t.is_identity = 1) THEN 1 ELSE 0 END;
    SET @sql = CASE WHEN @identity = 1 THEN N'SET IDENTITY_INSERT dbo.' + QUOTENAME(@target) + N' ON; ' ELSE N'' END
             + N'INSERT INTO dbo.' + QUOTENAME(@target) + N' (' + @collist + N') SELECT ' + @collist
             + N' FROM ' + @stage + N'; SET @n = @@ROWCOUNT; '
             + CASE WHEN @identity = 1 THEN N'SET IDENTITY_INSERT dbo.' + QUOTENAME(@target) + N' OFF;' ELSE N'' END;
    EXEC sp_executesql @sql, N'@n int OUTPUT', @n = @n OUTPUT;
    UPDATE #loaded SET rows_loaded = @n WHERE ord = @ord;
    IF @n <> (SELECT rows_in_file FROM #loaded WHERE ord = @ord)
        INSERT INTO #problems VALUES (@target, N'loaded ' + CAST(@n AS nvarchar(20)) + N' rows, expected '
            + CAST((SELECT rows_in_file FROM #loaded WHERE ord = @ord) AS nvarchar(20)));
    FETCH NEXT FROM load_cur INTO @ord, @target, @stage;
END
CLOSE load_cur;
DEALLOCATE load_cur;

-- 3. Every reference still lands -------------------------------------------------
-- Counted the same way in the source when this file was made. A reference that was
-- already dangling there is carried as it is; one that dangles only here is a failure.
DECLARE ref_cur CURSOR LOCAL FAST_FORWARD FOR
    SELECT child, col, parent, pcol FROM #refs;
OPEN ref_cur;
FETCH NEXT FROM ref_cur INTO @child, @col, @parent, @pcol;
WHILE @@FETCH_STATUS = 0
BEGIN
    IF COL_LENGTH(N'dbo.' + QUOTENAME(@child), @col) IS NOT NULL AND COL_LENGTH(N'dbo.' + QUOTENAME(@parent), @pcol) IS NOT NULL
    BEGIN
        SET @sql = N'SELECT @n = COUNT(*) FROM dbo.' + QUOTENAME(@child) + N' c WHERE c.' + QUOTENAME(@col)
                 + N' <> 0 AND NOT EXISTS (SELECT 1 FROM dbo.' + QUOTENAME(@parent) + N' p WHERE p.' + QUOTENAME(@pcol)
                 + N' = c.' + QUOTENAME(@col) + N');';
        EXEC sp_executesql @sql, N'@n int OUTPUT', @n = @n OUTPUT;
        UPDATE #refs SET dangling_here = @n WHERE child = @child AND col = @col AND parent = @parent AND pcol = @pcol;
    END
    FETCH NEXT FROM ref_cur INTO @child, @col, @parent, @pcol;
END
CLOSE ref_cur;
DEALLOCATE ref_cur;

INSERT INTO #problems
SELECT child, N'column ' + col + N': ' + CAST(dangling_here AS nvarchar(20)) + N' row(s) point at a ' + parent
     + N' row that is not there (the source had ' + CAST(dangling_in_source AS nvarchar(20))
     + N'). Run the earlier files of this batch first.'
  FROM #refs WHERE dangling_here > dangling_in_source;

IF EXISTS (SELECT 1 FROM #problems)
BEGIN
    ROLLBACK TRANSACTION;
    SELECT N'NOTHING WAS WRITTEN' AS result, target AS [table], problem FROM #problems ORDER BY target;
    RAISERROR(N'The loaded rows failed a check (see the Results tab). Everything was rolled back.', 16, 1);
    RETURN;
END

-- 4. What happened ----------------------------------------------------------------
SELECT target AS [table], rows_in_file, placeholders_removed AS api_placeholders_removed, rows_loaded
  FROM #loaded ORDER BY ord;

IF EXISTS (SELECT 1 FROM #refs WHERE dangling_in_source > 0 OR dangling_here > 0)
    SELECT child + N'.' + col AS [reference], parent + N'.' + pcol AS points_at,
           dangling_in_source, dangling_here, N'carried as it was in the source' AS note
      FROM #refs WHERE dangling_in_source > 0 OR dangling_here > 0 ORDER BY child, col;

IF EXISTS (SELECT 1 FROM #notes)
    SELECT target AS [table], note FROM #notes ORDER BY target;

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT N'committed';
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT N'rolled back - set @commit = 1 to write';
END

IF OBJECT_ID('tempdb..#m_tbl_bpi') IS NOT NULL DROP TABLE #m_tbl_bpi;
IF OBJECT_ID('tempdb..#m_tbl_bpi_general') IS NOT NULL DROP TABLE #m_tbl_bpi_general;
IF OBJECT_ID('tempdb..#m_tbl_bpi_accreditation') IS NOT NULL DROP TABLE #m_tbl_bpi_accreditation;
IF OBJECT_ID('tempdb..#m_tbl_bpi_address') IS NOT NULL DROP TABLE #m_tbl_bpi_address;
IF OBJECT_ID('tempdb..#m_tbl_bpi_branch_industries') IS NOT NULL DROP TABLE #m_tbl_bpi_branch_industries;
IF OBJECT_ID('tempdb..#m_tbl_bpi_contacts') IS NOT NULL DROP TABLE #m_tbl_bpi_contacts;
IF OBJECT_ID('tempdb..#m_tbl_bpi_entity') IS NOT NULL DROP TABLE #m_tbl_bpi_entity;
IF OBJECT_ID('tempdb..#m_tbl_bpi_finance') IS NOT NULL DROP TABLE #m_tbl_bpi_finance;
IF OBJECT_ID('tempdb..#m_tbl_bpi_history') IS NOT NULL DROP TABLE #m_tbl_bpi_history;
IF OBJECT_ID('tempdb..#m_tbl_bpi_industries') IS NOT NULL DROP TABLE #m_tbl_bpi_industries;
IF OBJECT_ID('tempdb..#m_tbl_bpi_items') IS NOT NULL DROP TABLE #m_tbl_bpi_items;
IF OBJECT_ID('tempdb..#plan')     IS NOT NULL DROP TABLE #plan;
IF OBJECT_ID('tempdb..#refs')     IS NOT NULL DROP TABLE #refs;
IF OBJECT_ID('tempdb..#problems') IS NOT NULL DROP TABLE #problems;
IF OBJECT_ID('tempdb..#notes')    IS NOT NULL DROP TABLE #notes;
IF OBJECT_ID('tempdb..#cols')     IS NOT NULL DROP TABLE #cols;
IF OBJECT_ID('tempdb..#loaded')   IS NOT NULL DROP TABLE #loaded;
