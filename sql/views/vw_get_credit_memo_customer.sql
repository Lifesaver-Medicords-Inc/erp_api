-- Customer picker for the Customer Credit Memo screen (§5.18).
--
-- Deliberately separate from vw_get_customer rather than a column added to it:
-- the two views identify a customer by DIFFERENT ids on purpose, and changing
-- vw_get_customer's meaning would reach Sales Invoice, which relies on it.
--
--   vw_get_customer            customer_id = a.based_id  (parent tbl_bpi.id)
--   vw_get_credit_memo_customer partner_id = a.id        (tbl_bpi_general.id)
--
-- Credit Memo needs the branch-level id because its own server-side guard,
-- partnerHasEntityType (credit_memo_service.go), verifies the partner actually
-- holds the "CUS" entity type - and tbl_bpi_entity keys on bpi_general_id, not
-- on the parent BPI. Feeding it the parent id made every customer Credit Memo
-- fail with "partner <n> is not registered as a Customer" even for customers
-- that ARE correctly registered (confirmed live: Bridge Inc is
-- tbl_bpi_general.id=40015 with a CUS row; its parent tbl_bpi.id=40026 has no
-- entity rows at all, and never should).
--
-- a.id is also what vw_get_supplier_trade already exposes as supplier_id, which
-- is why supplier Credit Memos were never affected - this view brings the
-- customer side onto that same convention.
-- CREATE OR ALTER (not plain CREATE): RunSQLMigrations re-runs every file in
-- sql/views on every API start and log.Fatal's on any error, so a plain CREATE
-- would take the whole API down on the second boot. Not ALTER either - that
-- fails the first time, on a database where this view doesn't exist yet.
CREATE OR ALTER VIEW [dbo].[vw_get_credit_memo_customer] AS
SELECT a.id AS partner_id,
    a.based_id AS parent_bpi_id,
    a.branch_name AS customer,
    a.customer_code,
    c.name AS payment_term,
    ISNULL(b.finance_tax_code, '') AS tax_code,
    e.location AS customer_address,
    bpi.tin
FROM tbl_bpi_general a
    -- Same branch-level link tbl_bpi_address uses (branch_id -> tbl_bpi_general.id);
    -- see vw_get_supplier_trade's own header note on finance_branch_id being the
    -- real FK here, not finance_based_id.
    LEFT JOIN tbl_bpi_finance b ON a.id = b.finance_branch_id
    LEFT JOIN tbl_setup_payment_terms c ON b.finance_payment_terms_id = c.id
    LEFT JOIN tbl_bpi_address e ON a.id = e.branch_id
    INNER JOIN tbl_bpi bpi ON a.based_id = bpi.id
    -- Only partners actually registered as customers can receive a customer
    -- Credit Memo, so the picker never offers one that would fail the guard.
    INNER JOIN tbl_bpi_entity be ON be.bpi_general_id = a.id
    INNER JOIN tbl_setup_bpi_entity ent ON ent.id = be.entity_id
    AND ent.code = 'CUS'
WHERE a.customer_code IS NOT NULL
    AND LTRIM(RTRIM(a.customer_code)) <> ''
