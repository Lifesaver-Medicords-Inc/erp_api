-- STEP 01 of 05 - setup lists - positions, users, company, chart of accounts, tax, and every item and BPI lookup
-- Run the numbered scripts in this folder in order, each on its own: 01, 02 ... 05.
-- Run after:  starting the API once against the NEW database, so it builds the tables
-- Run before: 02_items_2026-09-25.sql
-- Every step is a TRIAL RUN until you set @commit = 1; it must end by printing "committed".

-- Master data for a NEW database: setup lists - positions, users, company, chart of accounts, tax, and every item and BPI lookup
-- ============================================================================
-- The lists everything else points at: positions and users (with their existing passwords),
-- Company Setup, the chart of accounts and tax codes, payment terms, and every list Item
-- Entry and BPI pick from (item name, class, brand, unit, pump type and count, material...).
-- The chart of accounts keeps its ids on purpose: posting code refers to some accounts by id.
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
--   tbl_position (15)
--   tbl_position_access  (API placeholder rows removed only)
--   tbl_setup_users (16)
--   tbl_company (1)
--   tbl_setup_chart_class (5)
--   tbl_setup_chart_of_accounts (226)
--   tbl_setup_tax (2)
--   tbl_setup_tax_details (2)
--   tbl_setup_payment_terms (18)
--   tbl_setup_item_name (48)
--   tbl_setup_item_class (16)
--   tbl_setup_item_brand (28)
--   tbl_setup_item_unit_measurement (29)
--   tbl_setup_item_type (2)
--   tbl_setup_item_material (11)
--   tbl_setup_item_pump_type (8)
--   tbl_setup_item_pump_count (6)
--   tbl_setup_valuation_method (7)
--   tbl_inv_warehouse_usetype (5)
--   tbl_setup_application (5)
--   tbl_setup_bpi_entity (8)
--   tbl_setup_bpi_industries (39)
--   tbl_setup_ship_type (2)
--   tbl_setup_social (12)
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
IF OBJECT_ID('tempdb..#m_tbl_position') IS NOT NULL DROP TABLE #m_tbl_position;
IF OBJECT_ID('tempdb..#m_tbl_setup_users') IS NOT NULL DROP TABLE #m_tbl_setup_users;
IF OBJECT_ID('tempdb..#m_tbl_company') IS NOT NULL DROP TABLE #m_tbl_company;
IF OBJECT_ID('tempdb..#m_tbl_setup_chart_class') IS NOT NULL DROP TABLE #m_tbl_setup_chart_class;
IF OBJECT_ID('tempdb..#m_tbl_setup_chart_of_accounts') IS NOT NULL DROP TABLE #m_tbl_setup_chart_of_accounts;
IF OBJECT_ID('tempdb..#m_tbl_setup_tax') IS NOT NULL DROP TABLE #m_tbl_setup_tax;
IF OBJECT_ID('tempdb..#m_tbl_setup_tax_details') IS NOT NULL DROP TABLE #m_tbl_setup_tax_details;
IF OBJECT_ID('tempdb..#m_tbl_setup_payment_terms') IS NOT NULL DROP TABLE #m_tbl_setup_payment_terms;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_name') IS NOT NULL DROP TABLE #m_tbl_setup_item_name;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_class') IS NOT NULL DROP TABLE #m_tbl_setup_item_class;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_brand') IS NOT NULL DROP TABLE #m_tbl_setup_item_brand;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_unit_measurement') IS NOT NULL DROP TABLE #m_tbl_setup_item_unit_measurement;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_type') IS NOT NULL DROP TABLE #m_tbl_setup_item_type;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_material') IS NOT NULL DROP TABLE #m_tbl_setup_item_material;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_pump_type') IS NOT NULL DROP TABLE #m_tbl_setup_item_pump_type;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_pump_count') IS NOT NULL DROP TABLE #m_tbl_setup_item_pump_count;
IF OBJECT_ID('tempdb..#m_tbl_setup_valuation_method') IS NOT NULL DROP TABLE #m_tbl_setup_valuation_method;
IF OBJECT_ID('tempdb..#m_tbl_inv_warehouse_usetype') IS NOT NULL DROP TABLE #m_tbl_inv_warehouse_usetype;
IF OBJECT_ID('tempdb..#m_tbl_setup_application') IS NOT NULL DROP TABLE #m_tbl_setup_application;
IF OBJECT_ID('tempdb..#m_tbl_setup_bpi_entity') IS NOT NULL DROP TABLE #m_tbl_setup_bpi_entity;
IF OBJECT_ID('tempdb..#m_tbl_setup_bpi_industries') IS NOT NULL DROP TABLE #m_tbl_setup_bpi_industries;
IF OBJECT_ID('tempdb..#m_tbl_setup_ship_type') IS NOT NULL DROP TABLE #m_tbl_setup_ship_type;
IF OBJECT_ID('tempdb..#m_tbl_setup_social') IS NOT NULL DROP TABLE #m_tbl_setup_social;

-- What goes where. placeholder = the API's own rows, the only ones this may delete.
CREATE TABLE #plan (ord int NOT NULL PRIMARY KEY, target sysname NOT NULL, stage sysname NULL,
                    rows_in_file int NOT NULL, placeholder nvarchar(max) NULL);
INSERT INTO #plan (ord, target, stage, rows_in_file, placeholder) VALUES
    (10, N'tbl_position', N'#m_tbl_position', 15, N'code IN (N''DISPATCH'', N''CASHIER'') AND name IN (N''Dispatcher'', N''AR/AP Cashier'')'),
    (20, N'tbl_position_access', NULL, 0, N'position_id IN (SELECT id FROM dbo.tbl_position WHERE code IN (N''DISPATCH'', N''CASHIER'') AND name IN (N''Dispatcher'', N''AR/AP Cashier''))'),
    (30, N'tbl_setup_users', N'#m_tbl_setup_users', 16, NULL),
    (40, N'tbl_company', N'#m_tbl_company', 1, NULL),
    (50, N'tbl_setup_chart_class', N'#m_tbl_setup_chart_class', 5, NULL),
    (60, N'tbl_setup_chart_of_accounts', N'#m_tbl_setup_chart_of_accounts', 226, NULL),
    (70, N'tbl_setup_tax', N'#m_tbl_setup_tax', 2, NULL),
    (80, N'tbl_setup_tax_details', N'#m_tbl_setup_tax_details', 2, NULL),
    (90, N'tbl_setup_payment_terms', N'#m_tbl_setup_payment_terms', 18, NULL),
    (100, N'tbl_setup_item_name', N'#m_tbl_setup_item_name', 48, N'code = N''DEFAULT'' AND name = N''Default Item Name (Placeholder)'''),
    (110, N'tbl_setup_item_class', N'#m_tbl_setup_item_class', 16, N'code = N''DEFAULT'' AND name = N''Default (Placeholder)'''),
    (120, N'tbl_setup_item_brand', N'#m_tbl_setup_item_brand', 28, N'code = N''DEFAULT'' AND name = N''Default (Placeholder)'''),
    (130, N'tbl_setup_item_unit_measurement', N'#m_tbl_setup_item_unit_measurement', 29, N'code = N''UNIT'' AND name = N''Unit (Placeholder)'''),
    (140, N'tbl_setup_item_type', N'#m_tbl_setup_item_type', 2, NULL),
    (150, N'tbl_setup_item_material', N'#m_tbl_setup_item_material', 11, NULL),
    (160, N'tbl_setup_item_pump_type', N'#m_tbl_setup_item_pump_type', 8, NULL),
    (170, N'tbl_setup_item_pump_count', N'#m_tbl_setup_item_pump_count', 6, NULL),
    (180, N'tbl_setup_valuation_method', N'#m_tbl_setup_valuation_method', 7, NULL),
    (190, N'tbl_inv_warehouse_usetype', N'#m_tbl_inv_warehouse_usetype', 5, NULL),
    (1010, N'tbl_setup_application', N'#m_tbl_setup_application', 5, NULL),
    (1012, N'tbl_setup_bpi_entity', N'#m_tbl_setup_bpi_entity', 8, NULL),
    (1013, N'tbl_setup_bpi_industries', N'#m_tbl_setup_bpi_industries', 39, NULL),
    (1019, N'tbl_setup_ship_type', N'#m_tbl_setup_ship_type', 2, NULL),
    (1020, N'tbl_setup_social', N'#m_tbl_setup_social', 12, NULL);

-- References re-checked after loading, with how many already dangled in the source.
CREATE TABLE #refs (child sysname NOT NULL, col sysname NOT NULL, parent sysname NOT NULL, pcol sysname NOT NULL,
                    dangling_in_source int NOT NULL, dangling_here int NULL);
INSERT INTO #refs (child, col, parent, pcol, dangling_in_source) VALUES
    (N'tbl_setup_chart_of_accounts', N'class_id', N'tbl_setup_chart_class', N'id', 0),
    (N'tbl_setup_tax', N'coa_purchase_id', N'tbl_setup_chart_of_accounts', N'id', 2),
    (N'tbl_setup_tax', N'coa_sales_id', N'tbl_setup_chart_of_accounts', N'id', 1),
    (N'tbl_setup_tax_details', N'tax_code_id', N'tbl_setup_tax', N'id', 0),
    (N'tbl_setup_users', N'position_id', N'tbl_position', N'id', 0);

-- tbl_position (15 rows) ------------------------------------------------
CREATE TABLE #m_tbl_position ([id] bigint NULL, [code] nvarchar(100) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_position ([id], [code], [name]) VALUES
    (1, N'Admin', N'Admin'),
    (2, N'WH', N'Warehouse'),
    (3, N'SR', N'Sales Representatives'),
    (10002, N'PUR', N'Purchaser'),
    (10003, N'AR', N'Accounts Receivables'),
    (10004, N'AP', N'Accounts Payable'),
    (20002, N'ENG', N'Engineer'),
    (20003, N'SM', N'Sales Manager'),
    (20004, N'CBDO', N'Chief Business Development Officer'),
    (20017, N'CTR', N'Contractor'),
    (20018, N'EE', N'Electrical Engr.'),
    (20019, N'ME', N'Mechanical Engr.'),
    (20020, N'OWN', N'Owner'),
    (20021, N'PM', N'Project Manager'),
    (20022, N'SRP', N'Sales Representative');

-- tbl_setup_users (16 rows) ---------------------------------------------
CREATE TABLE #m_tbl_setup_users ([id] bigint NULL, [employee_id] nvarchar(256) NULL, [first_name] nvarchar(max) NULL, [last_name] nvarchar(max) NULL, [password] nvarchar(max) NULL, [position_id] bigint NULL, [department] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_users ([id], [employee_id], [first_name], [last_name], [password], [position_id], [department]) VALUES
    (1, N'IT-WD-1', N'Admin', N'Admin', N'$2a$10$Jkr2FeURBsZpFn5kBnvvc.z4us/Ililcqzg4hEO4K5OJ3ap2W6hZ.', 1, N'Admin'),
    (2, N'Admin--2', N'Gids', N'', N'$2a$10$vLnqBJe3LQ92RjriJR.3HuJrNu6aqlJN2pJO1HGso9lxUMP4WziLy', 1, N'Admin'),
    (3, N'Warehouse--3', N'Warehouse', N'', N'$2a$10$7FprwZztEgz77yfLotle0ev86c8TK/WpEjBFl.250ckqaJ.YAHZXG', 2, N'Warehouse'),
    (4, N'SAL-SR-4', N'Julie', N'CESTONA', N'$2a$10$66zclslOZ.3cv/1iKoLc..EretMFIy3yXtGgONoorw8vdj9706AyO', 20003, N'Sales'),
    (5, N'SAL-SR-5', N'JUANITO', N'SY', N'$2a$10$cqVMymRNr290uYYfu2iEV.bJuKeziRV6fCb7hqzB8Zld2Eryif6wa', 3, N'Sales'),
    (6, N'SAL-SR-6', N'L.', N'SALANGA', N'$2a$10$VJu96SlsrXb2PbZinElCFeV.DD.AyvtBq/7DIIs8SSig2FUWHW17a', 3, N'Sales'),
    (7, N'ADM-ADM-7', N'Cai', N'', N'$2a$10$jNwrnA3mKVEv1dBygq2FW.cFZu5Q.0p2vSdFfk0ZyXYbm3WTrhHg.', 1, N'Admin'),
    (8, N'SAL-SR-8', N'Rachelle', N'Robles', N'$2a$10$Jl7Zjx3cUyNbbRPwjLa0kOGnZee71ybFBSGzlpMXrnKhvKeZim3SW', 3, N'Sales'),
    (9, N'SAL-SR-9', N'April', N'Ponce', N'$2a$10$f1ZQcxA4i5SufK0NrEkGcO9.wbZlO30jaoic8bJMP/byHD7aaXTA2', 3, N'Sales'),
    (10, N'SAL-SR-10', N'Jezzryl', N'Azuelo', N'$2a$10$kPMNFj9m08xZ707Av0JoOuRJxdjvLgfJHQhqiRa2PFDDVvjWPc1C.', 3, N'Sales'),
    (11, N'SAL-SR-11', N'Lovely', N'Lopez', N'$2a$10$dsye0OqxPdZEFJvsDuWi1uvPyJxgGAJeioTgBIL.lQVDcxzFpDRhy', 3, N'Sales'),
    (12, N'SAL-SR-12', N'Jahziel', N'Gautane', N'$2a$10$xcYZ3TeB64khHDy4lxXLOuXKCh/udvxQzJibm2OzlhYtEsd1FwIMG', 3, N'Sales'),
    (13, N'PUR-PUR-13', N'Ayde', N'Soriano', N'$2a$10$7dROKOC7MsAc1Efs78UYX.orGu85EGUOX3JIhoxTIeZD3AYl3xZvu', 10002, N'Purchasing'),
    (14, N'ACC-AR-14', N'Pamela', N'Simplicio', N'$2a$10$9kKMZyNw16yGokfNHB6.2uVwwD966zbKDG5yURjbwFm5dhdMuITXa', 10003, N'Accounting'),
    (15, N'ACC-AP-15', N'Jonacel', N'Sibulan', N'$2a$10$oeRL8Yo9ArnEri1fFYswqeu80KzF1oOGaCZP654hngfOigAA3w6K6', 10004, N'Accounting'),
    (16, N'SAL-SR-16', N'Office', N'', N'$2a$10$cDVsZoRfsLij6J.8NrSRReVLhMgXOuKfefqr50drTdGpxdvRX4iIG', 3, N'Sales');

-- tbl_company (1 rows) -------------------------------------------------
CREATE TABLE #m_tbl_company ([id] bigint NULL, [company_code] nvarchar(max) NULL, [company_name] nvarchar(max) NULL, [legal_name] nvarchar(max) NULL, [trade_name] nvarchar(max) NULL, [business_type] nvarchar(max) NULL, [sec_registration_no] nvarchar(max) NULL, [dti_registration_no] nvarchar(max) NULL, [tin] nvarchar(max) NULL, [bir_branch_code] nvarchar(max) NULL, [rdo_code] nvarchar(max) NULL, [industry] nvarchar(max) NULL, [status] nvarchar(max) NULL, [is_head_office] bit NULL, [beg_bal] float NULL, [monthly_rate] float NULL, [currency_code] nvarchar(max) NULL, [markup_multiplier_price] float NULL, [start_fiscal_date] nvarchar(max) NULL, [end_fiscal_date] nvarchar(max) NULL, [inclusions_quotation_terms] nvarchar(max) NULL, [exclusions_quotation_terms] nvarchar(max) NULL, [term_and_conditions] nvarchar(max) NULL, [vat_rate_percent] float NULL, [restocking_fee_percent] float NULL, [cancellation_fee_percent] float NULL);
INSERT INTO #m_tbl_company ([id], [company_code], [company_name], [legal_name], [trade_name], [business_type], [sec_registration_no], [dti_registration_no], [tin], [bir_branch_code], [rdo_code], [industry], [status], [is_head_office], [beg_bal], [monthly_rate], [currency_code], [markup_multiplier_price], [start_fiscal_date], [end_fiscal_date], [inclusions_quotation_terms], [exclusions_quotation_terms], [term_and_conditions], [vat_rate_percent], [restocking_fee_percent], [cancellation_fee_percent]) VALUES
    (1, N'SMPC', N'Sunshine Multi-Plus Corporation', N'Sunshine Multi-Plus Corporation', N'Sunshine Multi-Plus Corporation', N'Distributor', N'A199709906', N'123456789', N'004-678-201-000', N'c-123456', N'RC-0001', N'Wholesale of Hardware, Plumbing, Heating Equipment', N'Active', 1, NULL, NULL, N'PHP', 1.1859999999999999e+000, N'1/1/2026 12:00:00 AM', N'12/31/2026 11:59:59 PM', N'1. Delivery at (PLACE) and supervision of unloading ' + NCHAR(13) + N'' + NCHAR(10) + N'     2. Testing and Commissioning at (PLACE)' + NCHAR(13) + N'' + NCHAR(10) + N'     - Validate compliance to Pre-Start-Up and Safety Checklist' + NCHAR(13) + N'' + NCHAR(10) + N'     - Configuration of Motor Voltage' + NCHAR(13) + N'' + NCHAR(10) + N'     - Termination of Load Side Wires (Motor to Controller)' + NCHAR(13) + N'' + NCHAR(10) + N'     - Placement of Sensors' + NCHAR(13) + N'' + NCHAR(10) + N'     - Prime the pump if necessary' + NCHAR(13) + N'' + NCHAR(10) + N'     - Initial and Final Testing' + NCHAR(13) + N'' + NCHAR(10) + N'     - Training and Orientation' + NCHAR(13) + N'' + NCHAR(10) + N'     3. For Testing and Commissioning, there is an allowance of 3 DAYS to be conducted during company regular working hours only. If you want more than the allowable and beyond working hours, additional charges will be applied.', N'1. Installation, piping works, and electrical' + NCHAR(13) + N'' + NCHAR(10) + N'     2. Factory/manufacturer''s test certificates or other similar documents' + NCHAR(13) + N'' + NCHAR(10) + N'     3. Supply of other services, materials, consumables, and accessories not mentioned in the offer' + NCHAR(13) + N'' + NCHAR(10) + N'     4. Permits and Bonds' + NCHAR(13) + N'' + NCHAR(10) + N'     5. Hauling of Equipment' + NCHAR(13) + N'' + NCHAR(10) + N'     6. Processing and Delivery Permits Fee to Site', N'PAYMENT TERMS: The Customer agrees to pay for the products according to the agreed payment terms with Sunshine Multi Plus Corporation. The agreed upon payment terms is CASH ON DELIVERY.' + NCHAR(13) + N'' + NCHAR(10) + N'     QUOTATION VALIDITY: Quote is valid 30 DAYS from date hereon; thereafter, it shall be subject to reconfirmation (On price validity, due to uncertainty, prices may change without prior notice).' + NCHAR(13) + N'' + NCHAR(10) + N'     AVAILABILITY OF STOCK(S) AND/OR SERVICE(S): 4-6 MONTHS upon receipt of purchase order and approval of setting plan (Due to uncertainty, delivery lead time may change without prior notice).' + NCHAR(13) + N'' + NCHAR(10) + N'     DELIVERY TERMS: Delivery terms are all ex-warehouse or as otherwise agreed to as evidenced by Sunshine Multi Plus Corporation''s order acknowledgement. Delivery instruction is WAREHOUSE TO SITE VIA SEA (w/o HAULING).' + NCHAR(13) + N'' + NCHAR(10) + N'     OTHER CHARGES, TITLE, RISK OF LOSS: Shipping, handling, hauling, testing, installation, and maintenance activities are additional charges unless expressed and indicated at the time of sale. Once the goods are delivered, the Customer assumes all risk of loss or damage.' + NCHAR(13) + N'' + NCHAR(10) + N'     STORAGE: It is the Customer''s responsibility to ensure the safekeeping of the products.' + NCHAR(13) + N'' + NCHAR(10) + N'     SALES RETURN / CANCELLATION POLICY: The exchange of purchased items within seven (7) days from the purchase date shall be allowed...' + NCHAR(13) + N'' + NCHAR(10) + N'     WARRANTY: Sunshine Multi Plus Corporation''s standard warranty shall cover all of the company''s products for ONE (1) YEAR against factory defects.' + NCHAR(13) + N'' + NCHAR(10) + N'     SERVICES: Sunshine Multi Plus Corporation will provide the services in a competent and skilled manner.' + NCHAR(13) + N'' + NCHAR(10) + N'     LIABILITY: Sunshine Multi Plus Corporation''s liability under this agreement will be limited to repair or replacement.', 12, NULL, NULL);

-- tbl_setup_chart_class (5 rows) ---------------------------------------
CREATE TABLE #m_tbl_setup_chart_class ([id] bigint NULL, [type] nvarchar(max) NULL, [code] nvarchar(256) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_chart_class ([id], [type], [code], [name]) VALUES
    (1, N'LIABILITY', N'20000', N'LIABILITY 2026'),
    (2, N'ASSET', N'10000', N'ASSET 10000'),
    (3, N'REVENUE', N'40000', N'REVENUE 2026'),
    (4, N'EQUITY', N'30000', N'EQUITY'),
    (5, N'EXPENSE', N'50000', N'EXPENSE');

-- tbl_setup_chart_of_accounts (226 rows) ---------------------------------
CREATE TABLE #m_tbl_setup_chart_of_accounts ([id] bigint NULL, [code] nvarchar(max) NULL, [name] nvarchar(max) NULL, [account_class] nvarchar(max) NULL, [class_id] bigint NULL, [group] nvarchar(max) NULL, [group_id] bigint NULL, [cash_flow_category] nvarchar(max) NULL, [liquidity_class] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_chart_of_accounts ([id], [code], [name], [account_class], [class_id], [group], [group_id], [cash_flow_category], [liquidity_class]) VALUES
    (40030, N'2000001-1', N'ACCOUNTS PAYABLE-TRADE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (50030, N'5000021-2', N'NON-TRADE INVENTORY', N'EXPENSE', 5, N'', 0, N'', N''),
    (70032, N'1000002-1', N'ACCOUNTS RECEIVABLES-TRADE PRIVATE', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70034, N'1000001-5', N'CASH ON HAND', N'ASSET 10000', 2, N'CASH', 70035, N'', N''),
    (70035, N'CASH', N'CASH', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70037, N'4000001-5', N'SALES', N'REVENUE 2026', 3, N'', 0, N'', N''),
    (70038, N'2000008-2', N'ADVANCE PAYABLE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70040, N'1000003-1', N'ACCOUNTS RECEIVABLES-TRADE GOVERNMENT', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70041, N'1000004', N'ADVANCES TO EMPLOYEE', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70042, N'1000004-1', N'ADV SALVACION LIZA RAMOS', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70043, N'1000004-2', N'ADV RONALIZA BETCO', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70044, N'1000004-3', N'ADV PAMELA SIMPLICIO', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70045, N'1000004-4', N'ADV AYDE SORIANO', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70046, N'1000004-5', N'ADV RONA CASILAN', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70047, N'1000004-6', N'ADV EFREN ESPALDON', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70048, N'1000004-7', N'ADV MARIO MORA', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70049, N'1000004-8', N'ADV ROLANDO BALDIVINO', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70050, N'1000004-9', N'ADV ROLANTE NAQUILA', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70051, N'1000004-10', N'ADV JULIE CRISTY CESTONA', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70052, N'1000004-11', N'ADV RIZALINA GAJISAN', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70053, N'1000004-12', N'ADV VIDELIO CALAGUI', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70054, N'1000004-13', N'ADV ROMEL LUBIANO', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70055, N'1000004-14', N'ADV MARVEN LUZON', N'ASSET 10000', 2, N'ADVANCES TO EMPLOYEE', 70041, N'', N''),
    (70056, N'1000005-1', N'MERCHANDISE INVENTORY', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70057, N'1000006-1', N'VAT INPUT', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70058, N'1000007-1', N'CREDITABLE WITHHOLDING TAX', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70059, N'1000007-2', N'CREDITABLE WITHHOLDING TAX (PENDING)', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70060, N'1000008-1', N'REALTY UNITS', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70061, N'1000009-1', N'FURNITURE & FIXTURE', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70062, N'1000010-1', N'OFFICE EQUIPTMENT', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70063, N'1000011-1', N'VEHICLE & DELIVERY EQUIPMENT', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70064, N'1000012-1', N'ACCUM.DEP-REALTY UNITS', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70065, N'1000012-2', N'ACCUM.DEP-FUR & FIXTURES', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70066, N'1000012-3', N'ACCUM.DEP-OFFICE EQUIPMENT', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70067, N'1000012-4', N'ACCUM.DEP-VEHICLES', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70068, N'1000013-1', N'GENERAL INVENTORY', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70069, N'1000014-1', N'PETTY CASH FUND', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70070, N'1000014-4', N'PETTY CASH FUND (PENDING)', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70071, N'1000015-1', N'AR UNAPPLIED', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70072, N'1000016-1', N'FOR TESTING ITEMS', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70073, N'1000018', N'SOFTWARE SYSTEM', N'ASSET 10000', 2, N'', 0, N'', N''),
    (70074, N'1000018-01', N'Q-ERP', N'ASSET 10000', 2, N'SOFTWARE SYSTEM', 70073, N'', N''),
    (70075, N'1000018-02', N'WEBSITE', N'ASSET 10000', 2, N'SOFTWARE SYSTEM', 70073, N'', N''),
    (70076, N'1000001-1', N'CHINA BANK', N'ASSET 10000', 2, N'CASH', 70035, N'', N''),
    (70077, N'1000001-2', N'METROBANK 1', N'ASSET 10000', 2, N'CASH', 70035, N'', N''),
    (70078, N'1000001-2.1', N'METRO BANK 2', N'ASSET 10000', 2, N'CASH', 70035, N'', N''),
    (70079, N'1000001-3', N'BANCO DE ORO 1', N'ASSET 10000', 2, N'CASH', 70035, N'', N''),
    (70080, N'1000001-3.1', N'BANCO DE ORO 2', N'ASSET 10000', 2, N'CASH', 70035, N'', N''),
    (70081, N'1000001-4', N'PETTY CASH', N'ASSET 10000', 2, N'CASH', 70035, N'', N''),
    (70082, N'1000001-6', N'EAST WEST BANK', N'ASSET 10000', 2, N'CASH', 70035, N'', N''),
    (70083, N'1000001-6.1', N'EAST WEST BANK 2', N'ASSET 10000', 2, N'CASH', 70035, N'', N''),
    (70084, N'1000001-7', N'SECURITY BANK', N'ASSET 10000', 2, N'CASH', 70035, N'', N''),
    (70085, N'2000001-2', N'ACCRUED EXPENSE PAYABLE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70086, N'2000002-1', N'PHILHEALTH PREMIUM PAYABLE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70087, N'2000003-1', N'HDMF PREMIUM PAYABLE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70088, N'2000003-2', N'PAG-IBIG LOAN PAYABLE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70089, N'2000003-3', N'HDMF LOAN PAYABLE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70090, N'2000004-1', N'SSS PREMIUM PAYABLE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70091, N'2000004-2', N'SSS LOAN PAYABLE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70092, N'2000005-1', N'VAT OUTPUT', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70093, N'2000005-2', N'VAT PAYABLE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70094, N'2000006-1', N'EXPANDED WITHHOLDING TAX', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70095, N'2000006-2', N'WITHHOLDING TAX PAYABLE-COMP', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70096, N'2000007-1', N'LOAN PAYABLE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70097, N'2000008-1', N'ACCOUNTS PAYABLE-NON-TRADE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70098, N'2000009-1', N'INCOME TAX PAYABLE', N'LIABILITY 2026', 1, N'', 0, N'', N''),
    (70099, N'3000001-1', N'RETAINED EARNINGS', N'EQUITY', 4, N'', 0, N'', N''),
    (70100, N'3000001-2', N'RESTRICTED RETAINED EARNINGS', N'EQUITY', 4, N'', 0, N'', N''),
    (70101, N'3000002-1', N'SHARE CAPITAL', N'EQUITY', 4, N'', 0, N'', N''),
    (70102, N'14000014-2', N'PETTY CASH GAIN', N'REVENUE 2026', 3, N'', 0, N'', N''),
    (70103, N'14000015-1', N'FOREIGN EXCHANGE GAIN/LOSS', N'REVENUE 2026', 3, N'', 0, N'', N''),
    (70104, N'4000001-1', N'SALES-VATABLE', N'REVENUE 2026', 3, N'', 0, N'', N''),
    (70105, N'4000001-2', N'SALES-GOVERNMENT', N'REVENUE 2026', 3, N'', 0, N'', N''),
    (70106, N'4000001-3', N'SALES-ZERO RATED', N'REVENUE 2026', 3, N'', 0, N'', N''),
    (70107, N'4000001-4', N'SALES-DISCOUNT', N'REVENUE 2026', 3, N'', 0, N'', N''),
    (70108, N'4000006-1', N'CHARGES DISCOUNT', N'REVENUE 2026', 3, N'', 0, N'', N''),
    (70109, N'4000006-2', N'PURCHASE DISCOUNT', N'REVENUE 2026', 3, N'', 0, N'', N''),
    (70110, N'4000007-1', N'MISCELLANEOUS INCOME', N'REVENUE 2026', 3, N'', 0, N'', N''),
    (70111, N'4000008-1', N'INVENTORY GAIN OR LOSS', N'REVENUE 2026', 3, N'', 0, N'', N''),
    (70112, N'15000014-3', N'PETTY CASH LOSS', N'EXPENSE', 5, N'', 0, N'', N''),
    (70113, N'5000001-1', N'ADVERTISING & PROMOTION', N'EXPENSE', 5, N'', 0, N'', N''),
    (70114, N'5000002-1', N'COMMISSION', N'EXPENSE', 5, N'', 0, N'', N''),
    (70115, N'5000003', N'COMMUNICATION, LIGHT & WATER', N'EXPENSE', 5, N'', 0, N'', N''),
    (70116, N'5000003-1', N'MERALCO 343612970101', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70117, N'5000003-2', N'GLOBE 917-6746088', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70118, N'5000003-3', N'PLDT 832-9723', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70119, N'5000003-4', N'GLOBE 917-8966066', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70120, N'5000003-5', N'GLOBE 917-8966071', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70121, N'5000003-6', N'GLOBE 917-8966070', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70122, N'5000003-7', N'SUN 922-8966067', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70123, N'5000003-8', N'SUN 922-8966071', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70124, N'5000003-9', N'SUN 922-8966074', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70125, N'5000003-10', N'SUN 922-8966065', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70126, N'5000003-11', N'SUN 922-8966070', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70127, N'5000003-12', N'SUN 922-8966066', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70128, N'5000003-13', N'MAYNILAD BILL', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70129, N'5000003-14', N'PLDT 247-7015', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70130, N'5000003-15', N'PLDT 247-7016', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70131, N'5000003-16', N'PLDT 247-7017', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70132, N'5000003-17', N'PLDT 245-0287', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70133, N'5000003-18', N'PLDT 245-0277', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70134, N'5000003-19', N'PLDT 245-0279', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70135, N'5000003-20', N'PLDT 247-6895', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70136, N'5000003-21', N'MERALCO 343612930101', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70137, N'5000003-22', N'MERALCO 343613190101', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70138, N'5000003-23', N'CONVERGE ICT', N'EXPENSE', 5, N'COMMUNICATION, LIGHT & WATER', 70115, N'', N''),
    (70139, N'5000004', N'DEPRECIATION', N'EXPENSE', 5, N'', 0, N'', N''),
    (70140, N'5000004-01', N'REALTY UNIT', N'EXPENSE', 5, N'DEPRECIATION', 70139, N'', N''),
    (70141, N'5000004-02', N'FURNITURE & FIXTURES', N'EXPENSE', 5, N'DEPRECIATION', 70139, N'', N''),
    (70142, N'5000004-03', N'OFFICE EQUIPMENT', N'EXPENSE', 5, N'DEPRECIATION', 70139, N'', N''),
    (70143, N'5000004-04', N'VEHICLES', N'EXPENSE', 5, N'DEPRECIATION', 70139, N'', N''),
    (70144, N'5000005', N'FRINGE BENEFITS (COLA, HOLIDAY PAY, AL)', N'EXPENSE', 5, N'', 0, N'', N''),
    (70145, N'5000005-01', N'COLA', N'EXPENSE', 5, N'FRINGE BENEFITS (COLA, HOLIDAY PAY, AL)', 70144, N'', N''),
    (70146, N'5000005-02', N'HOLIDAY PAY', N'EXPENSE', 5, N'FRINGE BENEFITS (COLA, HOLIDAY PAY, AL)', 70144, N'', N''),
    (70147, N'5000005-03', N'ALLOWANCE', N'EXPENSE', 5, N'FRINGE BENEFITS (COLA, HOLIDAY PAY, AL)', 70144, N'', N''),
    (70148, N'5000006', N'FUEL & OIL', N'EXPENSE', 5, N'', 0, N'', N''),
    (70149, N'5000006-1', N'DIESEL PZI-839', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70150, N'5000006-2', N'DIESEL TSH-475', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70151, N'5000006-3', N'DIESEL PMK-364', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70152, N'5000006-4', N'DIESEL UUF-288', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70153, N'5000006-5', N'DIESEL RLY-683', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70154, N'5000006-6', N'GASOLINE QY-1062', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70155, N'5000006-7', N'GASOLINE NM-5667', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70156, N'5000006-8', N'OTHERS', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70157, N'5000006-9', N'DIESEL CO-6589', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70158, N'5000006-10', N'DIESEL WKI-771', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70159, N'5000006-11', N'DIESEL ZTT-705', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70160, N'5000006-12', N'DIESEL LONG BROTHER', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70161, N'5000006-13', N'DIESEL FOR UJU-518', N'EXPENSE', 5, N'FUEL & OIL', 70148, N'', N''),
    (70162, N'5000007', N'OFFICE SUPPLIES', N'EXPENSE', 5, N'', 0, N'', N''),
    (70163, N'5000007-01', N'OFFICE SUPPLIES', N'EXPENSE', 5, N'OFFICE SUPPLIES', 70162, N'', N''),
    (70164, N'5000008', N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', N'EXPENSE', 5, N'', 0, N'', N''),
    (70165, N'5000008-2', N'CO-6589 ISUZU MUX', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70166, N'5000008-3', N'WKI-771 MITSUBISHI MONTERO', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70167, N'5000008-4', N'ZTT-705 TOYOTA FORTUNER', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70168, N'5000008-5', N'UJU-518 TALL BROTHER', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70169, N'5000008-6', N'RLY-683 BIG BROTHER', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70170, N'5000008-7', N'TSH-475 OLD BROTHER', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70171, N'5000008-8', N'RMK-364 CUTE', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70172, N'5000008-9', N'UUF-288 L-300', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70173, N'5000008-10', N'PZI-839 HINO', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70174, N'5000008-11', N'QY-1062', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70175, N'5000008-12', N'NM-5667', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70176, N'5000008-13', N'OFFICE EQUIPMENT', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70177, N'5000008-14', N'LONG BROTHER', N'EXPENSE', 5, N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', 70164, N'', N''),
    (70178, N'5000008-1', N'INTEREST AND PENALTIES', N'EXPENSE', 5, N'', 0, N'', N''),
    (70179, N'5000009', N'REPRESENTATION EXPENSES', N'EXPENSE', 5, N'', 0, N'', N''),
    (70180, N'5000009-01', N'MEAL ALLOWANCE', N'EXPENSE', 5, N'REPRESENTATION EXPENSES', 70179, N'', N''),
    (70181, N'5000009-1', N'PURCHASES', N'EXPENSE', 5, N'', 0, N'', N''),
    (70182, N'5000010', N'SALARIES AND ALLOWANCE', N'EXPENSE', 5, N'', 0, N'', N''),
    (70183, N'5000010-2', N'SAL RIZALINA GAJISAN', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70184, N'5000010-3', N'SAL VIDELIO CALAGUI', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70185, N'5000010-4', N'SAL RONALIZA BETCO', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70186, N'5000010-5', N'SAL PAMELA SIMPLICIO', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70187, N'5000010-6', N'SAL AYDE SORIANO', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70188, N'5000010-7', N'SAL RONA CASILAN', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70189, N'5000010-8', N'SAL EFREN ESPALDON', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70190, N'5000010-9', N'SAL MARIO MORA', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70191, N'5000010-10', N'SAL SALVACION LIZA RAMOS', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70192, N'5000010-11', N'SAL JULIE CRISTY CESTONA', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70193, N'5000010-12', N'SAL ROLANDO BALDIVINO', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70194, N'5000010-13', N'SAL ROLANTE NAQUILA', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70195, N'5000010-14', N'SAL ROMEL LUBIANO', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70196, N'5000010-15', N'SAL BRYAN GABRIEL TAN', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70197, N'5000010-16', N'SAL JOVY BRILLANTES', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70198, N'5000010-17', N'SAL ARTEMIO CABILA', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70199, N'5000010-18', N'SAL EDUARDO FERRANCO', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70200, N'5000010-19', N'SALARY FOR AGENCY', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70201, N'5000010-20', N'SAL CECILYN ORIAN', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70202, N'5000010-21', N'SALARIES & WAGES', N'EXPENSE', 5, N'SALARIES AND ALLOWANCE', 70182, N'', N''),
    (70203, N'5000010-1', N'MISCELLANEOUS', N'EXPENSE', 5, N'', 0, N'', N''),
    (70204, N'5000011', N'OVERTIME PAY', N'EXPENSE', 5, N'', 0, N'', N''),
    (70205, N'5000012', N'13TH MONTH PAY', N'EXPENSE', 5, N'', 0, N'', N''),
    (70206, N'5000012-1', N'OFFICERS BENEFIT', N'EXPENSE', 5, N'', 0, N'', N''),
    (70207, N'5000013', N'SSS, HDMF, PHILHEALTH', N'EXPENSE', 5, N'', 0, N'', N''),
    (70208, N'5000013-1', N'SSS PREMIUM EXPENSE', N'EXPENSE', 5, N'SSS, HDMF, PHILHEALTH', 70207, N'', N''),
    (70209, N'5000013-2', N'PHILHEALTH PREMIUM EXPENSE', N'EXPENSE', 5, N'SSS, HDMF, PHILHEALTH', 70207, N'', N''),
    (70210, N'5000013-3', N'HDMF EXPENSE', N'EXPENSE', 5, N'SSS, HDMF, PHILHEALTH', 70207, N'', N''),
    (70211, N'5000014', N'TAXES & LICENSES', N'EXPENSE', 5, N'', 0, N'', N''),
    (70212, N'5000014-01', N'TAXES & LICENSES', N'EXPENSE', 5, N'TAXES & LICENSES', 70211, N'', N''),
    (70213, N'5000015', N'FINES/ PENALTIES', N'EXPENSE', 5, N'', 0, N'', N''),
    (70214, N'5000015-1', N'CHECKPOINT', N'EXPENSE', 5, N'FINES/ PENALTIES', 70213, N'', N''),
    (70215, N'5000015-2', N'FINES & PENALTIES', N'EXPENSE', 5, N'FINES/ PENALTIES', 70213, N'', N''),
    (70216, N'5000016', N'TRANSPORTATION, TOLL, & PARKING', N'EXPENSE', 5, N'', 0, N'', N''),
    (70217, N'5000016-01', N'TRANSPORTATION', N'EXPENSE', 5, N'TRANSPORTATION, TOLL, & PARKING', 70216, N'', N''),
    (70218, N'5000016-02', N'TOLL FEE', N'EXPENSE', 5, N'TRANSPORTATION, TOLL, & PARKING', 70216, N'', N''),
    (70219, N'5000016-03', N'PARKING FOR UUF-288', N'EXPENSE', 5, N'TRANSPORTATION, TOLL, & PARKING', 70216, N'', N''),
    (70220, N'5000016-04', N'PARKING FOR RMK-364', N'EXPENSE', 5, N'TRANSPORTATION, TOLL, & PARKING', 70216, N'', N''),
    (70221, N'5000016-05', N'PARKING FOR PZI-839', N'EXPENSE', 5, N'TRANSPORTATION, TOLL, & PARKING', 70216, N'', N''),
    (70222, N'5000016-06', N'PARKING FOR QY-1062', N'EXPENSE', 5, N'TRANSPORTATION, TOLL, & PARKING', 70216, N'', N''),
    (70223, N'5000016-07', N'PARKING FOR NM-5667', N'EXPENSE', 5, N'TRANSPORTATION, TOLL, & PARKING', 70216, N'', N''),
    (70224, N'5000016-08', N'PARKING FOR WKI-771', N'EXPENSE', 5, N'TRANSPORTATION, TOLL, & PARKING', 70216, N'', N''),
    (70225, N'5000016-09', N'PARKING FOR ISUZU MUX', N'EXPENSE', 5, N'TRANSPORTATION, TOLL, & PARKING', 70216, N'', N''),
    (70226, N'5000017', N'VEHICLE REGISTRATION FEES', N'EXPENSE', 5, N'', 0, N'', N''),
    (70227, N'5000017-01', N'LTO REG WKI-771', N'EXPENSE', 5, N'VEHICLE REGISTRATION FEES', 70226, N'', N''),
    (70228, N'5000017-02', N'LTO REG QY-1062', N'EXPENSE', 5, N'VEHICLE REGISTRATION FEES', 70226, N'', N''),
    (70229, N'5000017-03', N'LTO REG RLY-683', N'EXPENSE', 5, N'VEHICLE REGISTRATION FEES', 70226, N'', N''),
    (70230, N'5000017-04', N'LTO REG RMK-364', N'EXPENSE', 5, N'VEHICLE REGISTRATION FEES', 70226, N'', N''),
    (70231, N'5000017-05', N'LTO REG TSH-475', N'EXPENSE', 5, N'VEHICLE REGISTRATION FEES', 70226, N'', N''),
    (70232, N'5000017-06', N'LTO REG NM-5667', N'EXPENSE', 5, N'VEHICLE REGISTRATION FEES', 70226, N'', N''),
    (70233, N'5000017-07', N'LTO REG UJU-518', N'EXPENSE', 5, N'VEHICLE REGISTRATION FEES', 70226, N'', N''),
    (70234, N'5000017-08', N'LTO REG UUF-288', N'EXPENSE', 5, N'VEHICLE REGISTRATION FEES', 70226, N'', N''),
    (70235, N'5000017-09', N'LTO REG PZI-839', N'EXPENSE', 5, N'VEHICLE REGISTRATION FEES', 70226, N'', N''),
    (70236, N'5000018', N'SHIPPING, FREIGHT HANDLING, BROKERS FEE', N'EXPENSE', 5, N'', 0, N'', N''),
    (70237, N'5000018-1', N'SHIPPING', N'EXPENSE', 5, N'SHIPPING, FREIGHT HANDLING, BROKERS FEE', 70236, N'', N''),
    (70238, N'5000018-2', N'FREIGHT AND HANDLING', N'EXPENSE', 5, N'SHIPPING, FREIGHT HANDLING, BROKERS FEE', 70236, N'', N''),
    (70239, N'5000018-3', N'WAREHOUSING', N'EXPENSE', 5, N'SHIPPING, FREIGHT HANDLING, BROKERS FEE', 70236, N'', N''),
    (70240, N'5000018-4', N'BROKER''S FEE', N'EXPENSE', 5, N'SHIPPING, FREIGHT HANDLING, BROKERS FEE', 70236, N'', N''),
    (70241, N'5000019', N'INSURANCE', N'EXPENSE', 5, N'', 0, N'', N''),
    (70242, N'5000019-2', N'WKI-771', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70243, N'5000019-3', N'CO-6589', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70244, N'5000019-4', N'ZTT-705', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70245, N'5000019-5', N'RMK-364', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70246, N'5000019-6', N'RLY-683', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70247, N'5000019-7', N'TSH-475', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70248, N'5000019-8', N'PZI-839', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70249, N'5000019-9', N'UJU-518', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70250, N'5000019-10', N'UUF-288 L-300', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70251, N'5000019-11', N'6750 CONTRACTORS ALL RISK', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70252, N'5000019-12', N'LONG BROTHER', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70253, N'5000019-13', N'2017 VOLVO V60CC D4', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70254, N'5000019-14', N'NM-5667', N'EXPENSE', 5, N'INSURANCE', 70241, N'', N''),
    (70255, N'5000019-1', N'BANK CHARGES', N'EXPENSE', 5, N'', 0, N'', N''),
    (70256, N'5000020-1', N'INVENTORY LOSS', N'EXPENSE', 5, N'', 0, N'', N''),
    (70257, N'5000021-1', N'TAX OTHERS', N'EXPENSE', 5, N'', 0, N'', N''),
    (70258, N'5000022-1', N'JANITORIAL AND MESSENGERIAL EXPENSE', N'EXPENSE', 5, N'', 0, N'', N'');

-- tbl_setup_tax (2 rows) -----------------------------------------------
CREATE TABLE #m_tbl_setup_tax ([id] bigint NULL, [code] nvarchar(100) NULL, [tax_desc] nvarchar(max) NULL, [input_tax_creditable] bit NULL, [coa_sales_id] bigint NULL, [coa_purchase_id] bigint NULL, [remarks] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_tax ([id], [code], [tax_desc], [input_tax_creditable], [coa_sales_id], [coa_purchase_id], [remarks]) VALUES
    (4, N'S1', N'S1', 0, 0, 1, N''),
    (5, N'VAT 12%', N'VAT 12%', 0, 70039, 1, N'VAT 12%');

-- tbl_setup_tax_details (2 rows) ---------------------------------------
CREATE TABLE #m_tbl_setup_tax_details ([id] bigint NULL, [tax_code_id] bigint NULL, [valid_from] nvarchar(max) NULL, [valid_to] nvarchar(max) NULL, [tax_rate] float NULL);
INSERT INTO #m_tbl_setup_tax_details ([id], [tax_code_id], [valid_from], [valid_to], [tax_rate]) VALUES
    (3, 4, N'01/01/2026', N'12/31/2026', 12),
    (4, 5, N'01/01/2026', N'12/31/2027', 12);

-- tbl_setup_payment_terms (18 rows) -------------------------------------
CREATE TABLE #m_tbl_setup_payment_terms ([id] bigint NULL, [code] nvarchar(100) NULL, [name] nvarchar(max) NULL, [is_selected] bit NULL);
INSERT INTO #m_tbl_setup_payment_terms ([id], [code], [name], [is_selected]) VALUES
    (1, N'CASH', N'CASH', 0),
    (4, N'GCASH', N'GCASH', NULL),
    (21, N'Cph', N'Coins.ph', NULL),
    (23, N'MAYA ', N'PayMaya', NULL),
    (25, N'BDO', N'Bangko De Oro', 0),
    (26, N'DP50', N'Downpayment 50%', NULL),
    (27, N'DP25', N'Downpayment 25%', NULL),
    (28, N'DP75', N'Downpayment 75%', NULL),
    (29, N'COD', N'Cash on Delivery', NULL),
    (30, N'COD-C', N'Cash on Delivery - Cheque', NULL),
    (31, N'NET15', N'Net 15', NULL),
    (32, N'NET30', N'Net 30', NULL),
    (33, N'NET60', N'Net 60', NULL),
    (34, N'PDC30', N'Post Dated Cheque 30', NULL),
    (35, N'PDC45', N'Post Dated Cheque 45', NULL),
    (36, N'PDC60', N'Post Dated Cheque 60', NULL),
    (37, N'50-50', N'50%-50%', NULL),
    (38, N'30-20-50', N'30%- 20%- 50%', NULL);

-- tbl_setup_item_name (48 rows) -----------------------------------------
CREATE TABLE #m_tbl_setup_item_name ([id] bigint NULL, [code] nvarchar(256) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_item_name ([id], [code], [name]) VALUES
    (1, N'ARV', N'AIR RELEASE VALVE'),
    (2, N'BV', N'BALL VALVE'),
    (3, N'BT', N'BLADDER TANK'),
    (4, N'BFV', N'BUTTERFLY VALVE'),
    (5, N'CP', N'COMMON PACKAGE'),
    (6, N'CTRL', N'CONTROLLER'),
    (7, N'DCH', N'DISCHARGE COMMON HEADER'),
    (8, N'DRB', N'DISCHARGE RUBBER BELLOW'),
    (9, N'FS', N'FLOAT SWITCH'),
    (10, N'FM', N'FLOW METER'),
    (11, N'GV', N'GATE VALVE'),
    (12, N'LLC', N'LIQUID LEVEL CONTROL'),
    (13, N'PG', N'PRESSURE GAUGE'),
    (14, N'PRV', N'PRESSURE RELIEF VALVE'),
    (15, N'PS', N'PRESSURE SWITCH'),
    (16, N'PT', N'PRESSURE TRANSDUCER'),
    (17, N'PMP', N'PUMP'),
    (18, N'SCH', N'SUCTION COMMON HEADER'),
    (19, N'SRB', N'SUCTION RUBBER BELLOW'),
    (20, N'SCV', N'SUCTION CHECK VALVE'),
    (21, N'VFD', N'VARIABLE FREQUENCY DRIVE'),
    (22, N'WC', N'WASTE CONE'),
    (23, N'WM', N'WATER METER'),
    (24, N'WMAT', N'WIRING MATERIALS'),
    (26, N'SER', N'SERVICE'),
    (27, N'CONS', N'CONSUMABLES'),
    (28, N'ELEC', N'ELECTRICAL'),
    (29, N'FLV', N'FLOAT VALVE'),
    (30, N'FTV', N'FOOT VALVE'),
    (31, N'HDW', N'HARDWARE'),
    (32, N'OTH', N'OTHER'),
    (33, N'PNT', N'PAINT'),
    (34, N'PMPP', N'PUMP PART'),
    (35, N'SUP', N'SUPPPLIES'),
    (36, N'TAI', N'TOOLS AND INSTRUMENTS'),
    (37, N'TEST 1', N'TEST ITEM 1'),
    (62, N'AR', N'ALTERNATING RELAY'),
    (63, N'MC', N'MAGNETIC CONTACTOR'),
    (64, N'ME', N'METAL ENCLOSURE'),
    (65, N'MCB', N'MINIATURE CIRCUIT BREAKER'),
    (66, N'MCCB', N'MOLDED CASE CIRCUIT BREAKER'),
    (67, N'OUV', N'O/U VOLTAGE'),
    (68, N'PL', N'PILOT LIGHT'),
    (69, N'PB', N'PUSH BUTTON'),
    (70, N'SS', N'SELECTOR SWITCH'),
    (71, N'TB', N'TERMINAL BLOCK'),
    (72, N'TL', N'TERMINAL LUG'),
    (73, N'TOR', N'THERMAL OVERLOAD RELAY');

-- tbl_setup_item_class (16 rows) ----------------------------------------
CREATE TABLE #m_tbl_setup_item_class ([id] bigint NULL, [code] nvarchar(256) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_item_class ([id], [code], [name]) VALUES
    (1, N'SUP', N'SUPPLIES (CLEANING, SCHOOL, ETC.)'),
    (2, N'HW', N'HARDWARE'),
    (3, N'IND', N'INDUSTRIAL EQUIPMENT'),
    (4, N'PMP', N'PUMP'),
    (5, N'CON', N'CONSTRUCTION'),
    (6, N'CTI', N'COMPANY TOOLS AND INSTRUMENTS'),
    (7, N'EXP', N'EXPENSE (EMPTY SPECS)'),
    (8, N'SER', N'SERVICE'),
    (9, N'ELEC', N'ELECTRICAL'),
    (10, N'TAI', N'TOOLS AND INSTRUMENTS'),
    (11, N'IMP', N'IMPORT PURCHASE'),
    (12, N'LOC', N'LOCAL PURCHASE'),
    (13, N'OFF', N'OFFICE'),
    (14, N'PMB', N'PLUMBING'),
    (15, N'WM', N'WATER METER'),
    (16, N'PMPP', N'PUMP PARTS');

-- tbl_setup_item_brand (28 rows) ----------------------------------------
CREATE TABLE #m_tbl_setup_item_brand ([id] bigint NULL, [code] nvarchar(100) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_item_brand ([id], [code], [name]) VALUES
    (1, N'SHN', N'Shinmaywa'),
    (2, N'CLP', N'Calpeda'),
    (3, N'ACE', N'Acejet'),
    (4, N'DEL', N'Delta'),
    (5, N'LGI', N'Little Giant'),
    (6, N'ISP', N'Instapure'),
    (7, N'PTA', N'Pentair Aurora'),
    (8, N'FTI', N'Finish Thompson Inc.'),
    (10, N'WDE', N'Wede'),
    (11, N'DAN', N'Danfoss'),
    (12, N'DYNA', N'Dynaflo'),
    (13, N'ETN', N'Eaton'),
    (14, N'NB', N'No Brand'),
    (15, N'MAC', N'Mac 3'),
    (44, N'AND', N'ANDELI'),
    (45, N'ANLY', N'ANLY'),
    (46, N'CHS', N'CHS'),
    (47, N'DAL', N'Dalier'),
    (48, N'FUJ', N'Fuji'),
    (49, N'JKN', N'JKN Under'),
    (50, N'LS', N'LS'),
    (51, N'OMR', N'Omron'),
    (52, N'PHF', N'Philflex'),
    (53, N'SKD', N'Sakura Denso'),
    (54, N'SCHN', N'Schneider'),
    (55, N'SQD', N'Square D'),
    (56, N'WIX', N'Wixim'),
    (57, N'YKY', N'Yokoyama');

-- tbl_setup_item_unit_measurement (29 rows) -----------------------------
CREATE TABLE #m_tbl_setup_item_unit_measurement ([id] bigint NULL, [code] nvarchar(100) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_item_unit_measurement ([id], [code], [name]) VALUES
    (1, N'BAG', N'Bag'),
    (2, N'BKS', N'Books'),
    (3, N'BTL', N'Bottle'),
    (4, N'CAN', N'Can'),
    (5, N'CBY', N'Cubic Yard'),
    (6, N'DOZ.', N'Dozen'),
    (7, N'DRM', N'Drum'),
    (8, N'FEET', N'Feet'),
    (9, N'GAL', N'Gallon'),
    (10, N'GR.', N'Gross'),
    (11, N'KEG', N'Keg'),
    (12, N'KG', N'Kilogram'),
    (13, N'LM', N'Linear Meter'),
    (14, N'LOT', N'Lot'),
    (15, N'LTR', N'Liter'),
    (16, N'MTR', N'Meter'),
    (17, N'PACK', N'Pack'),
    (18, N'PAD', N'Pad'),
    (19, N'PAIL', N'Pail'),
    (20, N'PC', N'Piece'),
    (21, N'PRS', N'Pairs'),
    (22, N'REAM', N'Ream'),
    (23, N'ROLL', N'Roll'),
    (24, N'SET', N'Set'),
    (25, N'SHEET', N'Sheet'),
    (26, N'TANK', N'Tank'),
    (27, N'TIN', N'Tin'),
    (28, N'UNIT', N'Unit'),
    (29, N'QAT', N'Quart');

-- tbl_setup_item_type (2 rows) -----------------------------------------
CREATE TABLE #m_tbl_setup_item_type ([id] bigint NULL, [code] nvarchar(256) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_item_type ([id], [code], [name]) VALUES
    (1, N'TRD', N'TRADE'),
    (2, N'NTR', N'NON-TRADE');

-- tbl_setup_item_material (11 rows) -------------------------------------
CREATE TABLE #m_tbl_setup_item_material ([id] bigint NULL, [code] nvarchar(256) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_item_material ([id], [code], [name]) VALUES
    (1, N'STA', N'STAINLESS STEEL'),
    (2, N'AISL 304', N'STAINLESS STEEL 304'),
    (3, N'AISL 316', N'STAINLESS STEEL 316'),
    (4, N'DI', N'DUCTILE IRON'),
    (5, N'BI', N'BLACK IRON'),
    (6, N'GB', N'BRONZE'),
    (7, N'GG', N'CAST IRON'),
    (8, N'GI', N'GALVANIZED IRON'),
    (9, N'GO', N'BRASS'),
    (10, N'PPO-GF20', N'NORYL'),
    (11, N'FS', N'FORGED STEEL');

-- tbl_setup_item_pump_type (8 rows) ------------------------------------
CREATE TABLE #m_tbl_setup_item_pump_type ([id] bigint NULL, [code] nvarchar(256) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_item_pump_type ([id], [code], [name]) VALUES
    (1, N'CPS', N'CPS'),
    (2, N'ANO', N'ANO'),
    (3, N'BP', N'BOOSTER PUMP'),
    (4, N'DP', N'DRAINAGE PUMP'),
    (5, N'NUFP', N'NON-ULFM FIRE PUMP'),
    (6, N'NUJP', N'NON-ULFM JOCKEY PUMP'),
    (7, N'SP', N'SEWAGE PUMP'),
    (8, N'TP', N'TRANSFER PUMP');

-- tbl_setup_item_pump_count (6 rows) -----------------------------------
CREATE TABLE #m_tbl_setup_item_pump_count ([id] bigint NULL, [code] nvarchar(256) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_item_pump_count ([id], [code], [name]) VALUES
    (1, N'CPS', N'CPS'),
    (2, N'SMP', N'Simplex'),
    (3, N'DPX', N'Duplex'),
    (4, N'TPX', N'Triplex'),
    (5, N'QPX', N'Quadruplex'),
    (6, N'QPX2', N'Quintuplex');

-- tbl_setup_valuation_method (7 rows) ----------------------------------
CREATE TABLE #m_tbl_setup_valuation_method ([id] bigint NULL, [code] nvarchar(100) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_valuation_method ([id], [code], [name]) VALUES
    (4, N'DCF', N'Discounted Cash Flow'),
    (5, N'FIFO', N'First In First Out'),
    (6, N'LIFO', N'Last In First Out'),
    (7, N'WAC', N'Weighted Average Cost'),
    (8, N'MAC', N'Moving Average Cost'),
    (9, N'JIT', N'Just In Time'),
    (10, N'STD', N'Standard Costing');

-- tbl_inv_warehouse_usetype (5 rows) -----------------------------------
CREATE TABLE #m_tbl_inv_warehouse_usetype ([id] bigint NULL, [code] nvarchar(256) NULL, [name] nvarchar(max) NULL, [bg_color] nvarchar(max) NULL);
INSERT INTO #m_tbl_inv_warehouse_usetype ([id], [code], [name], [bg_color]) VALUES
    (1, N'DMG', N'DAMAGED ITEMS', NULL),
    (2, N'GEN', N'GENERAL', NULL),
    (3, N'OUT', N'OUTBOUND', NULL),
    (4, N'REC', N'RECEIVING', NULL),
    (5, N'SP', N'SPARE PARTS', NULL);

-- tbl_setup_application (5 rows) ---------------------------------------
CREATE TABLE #m_tbl_setup_application ([id] bigint NULL, [code] nvarchar(256) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_application ([id], [code], [name]) VALUES
    (10002, N'HVAC', N'HVAC'),
    (10003, N'DEWATERING', N'DEWATERING'),
    (10004, N'PLUMBING', N'PLUMBING'),
    (10005, N'SANITARY', N'SANITARY'),
    (10006, N'FIRE PROTECTION', N'FIRE PROTECTION');

-- tbl_setup_bpi_entity (8 rows) ----------------------------------------
CREATE TABLE #m_tbl_setup_bpi_entity ([id] bigint NULL, [code] nvarchar(100) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_bpi_entity ([id], [code], [name]) VALUES
    (1, N'SUP', N'Supplier'),
    (2, N'CUS', N'Customer'),
    (3, N'TEMP', N'TEMPORARY'),
    (4, N'AFF', N'Affiliated'),
    (5, N'BLK', N'Blacklisted'),
    (6, N'CLD', N'Closed'),
    (7, N'NAF', N'Non-Affiliated'),
    (8, N'TSP', N'Temporary Supplier');

-- tbl_setup_bpi_industries (39 rows) ------------------------------------
CREATE TABLE #m_tbl_setup_bpi_industries ([id] bigint NULL, [code] nvarchar(100) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_bpi_industries ([id], [code], [name]) VALUES
    (1, N'RET', N'Retail'),
    (3, N'AUTO', N'Automotive'),
    (4, N'AERO', N'Aerospace'),
    (6, N'EDU', N'Education'),
    (7, N'Energy & Utilities', N'Energy & Utilities'),
    (8, N'EM', N'Entertainment & Media'),
    (9, N'Finance & Banking', N'Finance & Banking'),
    (11, N'Hospitality & Tourism', N'Hospitality & Tourism'),
    (12, N'IT', N'Information Technology (IT)'),
    (13, N'INSU', N'Insurance'),
    (14, N'AGRI', N'Agriculture/ Irrigation'),
    (15, N'AQUA', N'Aquaculture'),
    (16, N'CHEM', N'Chemical'),
    (17, N'COLD', N'Cold Storage'),
    (18, N'CON', N'Construction'),
    (19, N'DATA', N'Data Centers'),
    (20, N'DOM', N'Domestic/ Residential'),
    (21, N'FILTER', N'Filtration'),
    (22, N'FIRE', N'Fire Systems'),
    (23, N'FBV', N'Food / Beverage'),
    (24, N'GRND', N'Groundwater'),
    (25, N'HC', N'Healthcare'),
    (26, N'HVAC', N'HVAC'),
    (27, N'IND', N'Industrial'),
    (28, N'ICE', N'Ice Plant'),
    (29, N'LAND', N'Landscaping'),
    (30, N'MRN', N'Marine'),
    (31, N'MINE', N'Mining'),
    (32, N'MUNI', N'Municipal'),
    (33, N'OIL', N'Oil / Gas'),
    (34, N'PHRM', N'Pharmaceuticals'),
    (35, N'PWR', N'Power Generation'),
    (36, N'POOL', N'Pools'),
    (37, N'SC', N'Semiconductors'),
    (38, N'SKCM', N'Skyscrapers/ Commercial'),
    (39, N'WT', N'Water Treatment'),
    (41, N'OTH', N'Others'),
    (43, N'HW', N'Hardware'),
    (45, N'TEST', N'TEST');

-- tbl_setup_ship_type (2 rows) -----------------------------------------
CREATE TABLE #m_tbl_setup_ship_type ([id] bigint NULL, [ship_name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_ship_type ([id], [ship_name]) VALUES
    (1, N'DELIVERY'),
    (2, N'PICK UP');

-- tbl_setup_social (12 rows) --------------------------------------------
CREATE TABLE #m_tbl_setup_social ([id] bigint NULL, [code] nvarchar(100) NULL, [name] nvarchar(max) NULL);
INSERT INTO #m_tbl_setup_social ([id], [code], [name]) VALUES
    (2, N'VBR', N'VIBER'),
    (3, N'EML', N'Email'),
    (4, N'FB', N'Facebook'),
    (5, N'IG', N'Instagram'),
    (6, N'INT', N'Internet Search'),
    (7, N'LNK', N'LinkedIn'),
    (8, N'REF', N'Referral'),
    (9, N'ST', N'Sales Team'),
    (10, N'TG', N'Telegram'),
    (11, N'TK', N'TikTok'),
    (12, N'WEB', N'Website'),
    (13, N'WSP', N'WhatsApp');

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

IF OBJECT_ID('tempdb..#m_tbl_position') IS NOT NULL DROP TABLE #m_tbl_position;
IF OBJECT_ID('tempdb..#m_tbl_setup_users') IS NOT NULL DROP TABLE #m_tbl_setup_users;
IF OBJECT_ID('tempdb..#m_tbl_company') IS NOT NULL DROP TABLE #m_tbl_company;
IF OBJECT_ID('tempdb..#m_tbl_setup_chart_class') IS NOT NULL DROP TABLE #m_tbl_setup_chart_class;
IF OBJECT_ID('tempdb..#m_tbl_setup_chart_of_accounts') IS NOT NULL DROP TABLE #m_tbl_setup_chart_of_accounts;
IF OBJECT_ID('tempdb..#m_tbl_setup_tax') IS NOT NULL DROP TABLE #m_tbl_setup_tax;
IF OBJECT_ID('tempdb..#m_tbl_setup_tax_details') IS NOT NULL DROP TABLE #m_tbl_setup_tax_details;
IF OBJECT_ID('tempdb..#m_tbl_setup_payment_terms') IS NOT NULL DROP TABLE #m_tbl_setup_payment_terms;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_name') IS NOT NULL DROP TABLE #m_tbl_setup_item_name;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_class') IS NOT NULL DROP TABLE #m_tbl_setup_item_class;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_brand') IS NOT NULL DROP TABLE #m_tbl_setup_item_brand;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_unit_measurement') IS NOT NULL DROP TABLE #m_tbl_setup_item_unit_measurement;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_type') IS NOT NULL DROP TABLE #m_tbl_setup_item_type;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_material') IS NOT NULL DROP TABLE #m_tbl_setup_item_material;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_pump_type') IS NOT NULL DROP TABLE #m_tbl_setup_item_pump_type;
IF OBJECT_ID('tempdb..#m_tbl_setup_item_pump_count') IS NOT NULL DROP TABLE #m_tbl_setup_item_pump_count;
IF OBJECT_ID('tempdb..#m_tbl_setup_valuation_method') IS NOT NULL DROP TABLE #m_tbl_setup_valuation_method;
IF OBJECT_ID('tempdb..#m_tbl_inv_warehouse_usetype') IS NOT NULL DROP TABLE #m_tbl_inv_warehouse_usetype;
IF OBJECT_ID('tempdb..#m_tbl_setup_application') IS NOT NULL DROP TABLE #m_tbl_setup_application;
IF OBJECT_ID('tempdb..#m_tbl_setup_bpi_entity') IS NOT NULL DROP TABLE #m_tbl_setup_bpi_entity;
IF OBJECT_ID('tempdb..#m_tbl_setup_bpi_industries') IS NOT NULL DROP TABLE #m_tbl_setup_bpi_industries;
IF OBJECT_ID('tempdb..#m_tbl_setup_ship_type') IS NOT NULL DROP TABLE #m_tbl_setup_ship_type;
IF OBJECT_ID('tempdb..#m_tbl_setup_social') IS NOT NULL DROP TABLE #m_tbl_setup_social;
IF OBJECT_ID('tempdb..#plan')     IS NOT NULL DROP TABLE #plan;
IF OBJECT_ID('tempdb..#refs')     IS NOT NULL DROP TABLE #refs;
IF OBJECT_ID('tempdb..#problems') IS NOT NULL DROP TABLE #problems;
IF OBJECT_ID('tempdb..#notes')    IS NOT NULL DROP TABLE #notes;
IF OBJECT_ID('tempdb..#cols')     IS NOT NULL DROP TABLE #cols;
IF OBJECT_ID('tempdb..#loaded')   IS NOT NULL DROP TABLE #loaded;
