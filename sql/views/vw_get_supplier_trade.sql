-- Bug fix (Invoice Receipt): the finance join used b.finance_id, which is
-- tbl_bpi_finance's OWN primary key (see bpi_finance_model.go - `FinanceID uint
-- gorm:"primarykey"`), not a foreign key back to the BPI. It only ever matched a
-- supplier by coincidence of the two tables' independent identity sequences
-- lining up, so payment_term came back NULL for effectively every real supplier -
-- confirmed live on "Calpeda" (BPI has PAYMENT TERMS=CASH, TAX CODE=S1 configured,
-- Invoice Receipt showed both blank).
--
-- The real FK is finance_branch_id, NOT finance_based_id (tried first, also
-- wrong - verified directly against the live table: Calpeda is
-- tbl_bpi_general.id=30008, and its actual finance row has
-- finance_branch_id=30008 but finance_based_id=30022, which belongs to a
-- different BPI entirely). tbl_bpi_finance follows the same "branch_id points at
-- tbl_bpi_general.id" convention already used by tbl_bpi_address.branch_id just
-- below (BPI records are modeled as "branches" under the hood - see
-- a.branch_name above) - finance_based_id is something else, not this link.
--
-- Also adds tax_code (a plain string on tbl_bpi_finance, not a lookup id) so
-- Invoice Receipt can look it up against Tax Setup and default cmb_tax_code the
-- same way payment_term defaults txt_payment_term.
ALTER VIEW [dbo].[vw_get_supplier_trade] AS
SELECT a.id AS supplier_id,
    a.branch_name AS supplier,
    a.supplier_code,
    UPPER(a.transaction_type) AS invoice_type,
    c.name AS payment_term,
    ISNULL(b.finance_tax_code, '') AS tax_code,
    'INVOICE TYPE' AS [type],
    e.location AS supplier_address,
    CASE
        WHEN ISNULL(SUM(d.company_overpayment_amount), 0) < 0.01 THEN 0
        ELSE ISNULL(SUM(d.company_overpayment_amount), 0)
    END AS overpayment_amount
FROM tbl_bpi_general a
    LEFT JOIN tbl_bpi_finance b ON a.id = b.finance_branch_id
    LEFT JOIN tbl_setup_payment_terms c ON b.finance_payment_terms_id = c.id
    LEFT JOIN tbl_accounting_bpi_overpayment d ON a.id = d.bpi_id
    LEFT JOIN tbl_bpi_address e ON a.id = e.branch_id
GROUP BY a.id,
    a.branch_name,
    a.supplier_code,
    a.transaction_type,
    c.name,
    b.finance_tax_code,
    e.location;