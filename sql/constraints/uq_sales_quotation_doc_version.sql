-- Sales_Quotation_Bug_Report_2026-08-03.md #13 - nothing enforced uniqueness of
-- (document_no, version_no, sub_version_no), so a stale client-side list or a
-- genuinely concurrent request could create two rows for what should be one
-- specific version. The application-layer check in CreateSalesQuotation/
-- CreateSalesProject (commit b3a859f) closes this for normal sequential use,
-- but a check-then-insert isn't atomic - two requests landing at the same
-- instant could both pass the check before either insert lands. This index is
-- the actual backstop that closes that race completely.
--
-- Prerequisite, already applied live: document_no/version_no/sub_version_no
-- were NVARCHAR(MAX) (GORM's default for a plain Go string field) - SQL Server
-- refuses to let a MAX-length column be an index key at all. Narrowed to
-- NVARCHAR(50)/NVARCHAR(20)/NVARCHAR(20) first (real data's own max lengths
-- were 7/1/1 chars at the time - still generous room for the full documented
-- format, "FQ#YYYY-nnnn-v#.#"). See models/sales_quotation_model.go's own
-- gorm:"size:..." tags on these fields, added alongside this.
--
-- Deliberately NOT applied to z_tbl_trans_sales_quotation_at (the audit
-- table), which shares the same three column names via the same embedded
-- SalesQuotationContent struct but legitimately holds many rows per
-- document/version - a unique constraint there would break the audit trail.
--
-- Verified before creating: 0 existing duplicate (document_no, version_no,
-- sub_version_no) combinations in the live data. Verified after creating,
-- in a rolled-back transaction: a genuine duplicate insert is rejected, and
-- New Version (same document_no, different version_no) still succeeds.

ALTER TABLE tbl_trans_sales_quotation ALTER COLUMN document_no NVARCHAR(50) NULL;
ALTER TABLE tbl_trans_sales_quotation ALTER COLUMN version_no NVARCHAR(20) NULL;
ALTER TABLE tbl_trans_sales_quotation ALTER COLUMN sub_version_no NVARCHAR(20) NULL;

ALTER TABLE z_tbl_trans_sales_quotation_at ALTER COLUMN document_no NVARCHAR(50) NULL;
ALTER TABLE z_tbl_trans_sales_quotation_at ALTER COLUMN version_no NVARCHAR(20) NULL;
ALTER TABLE z_tbl_trans_sales_quotation_at ALTER COLUMN sub_version_no NVARCHAR(20) NULL;

CREATE UNIQUE INDEX UQ_sales_quotation_doc_version
    ON tbl_trans_sales_quotation (document_no, version_no, sub_version_no);
