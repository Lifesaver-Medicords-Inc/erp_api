ALTER VIEW [dbo].[vw_get_engineering_redbox_quotation_list] AS
SELECT sq.id AS id,
    bpi.name AS client_name,
    sq.document_no AS sales_quotation,
    so.status AS status,
    sq.project_name AS project_name,
    sq.created_by AS sales_executive,
    -- §9's own text: this remark is specifically "NEEDS WIRING" where wiring was
    -- ticked, not the quote's free-text purpose (sq.purpose, the previous value -
    -- correct by accident on quotes that happened to use purpose for that, wrong
    -- on any that didn't).
    CASE WHEN EXISTS (
        SELECT 1
        FROM dbo.tbl_trans_sales_project_wiring w
        INNER JOIN tbl_trans_sales_project_item_set spis ON w.based_id = spis.item_set_id
        WHERE spis.based_id = sq.id
    ) THEN 'NEEDS WIRING' ELSE '' END AS remark,
    sq.requested_engr_id AS requested_engr_id,
    -- DATE REQUESTED / requester name for the engineering Sales Quotation List grid
    -- (col_date_requested binds requested_for_engr_date). Were not selected before,
    -- so that column rendered blank even once rows appeared.
    sq.requested_for_engr_date AS requested_for_engr_date,
    sq.requested_engr_name AS requested_engr_name
FROM tbl_trans_sales_quotation sq
    -- LEFT JOIN (was INNER JOIN): a quotation whose customer record is
    -- missing/soft-deleted used to vanish from this list entirely instead
    -- of just showing a blank client name.
    LEFT JOIN tbl_bpi bpi ON sq.customer_id = bpi.id
    -- OUTER APPLY TOP 1 (was INNER JOIN on document_no): the opportunity
    -- table isn't unique per document_no, so the plain join fanned out
    -- duplicate rows for every matching opportunity. Take the most recent
    -- opportunity per document instead of joining to all of them.
    OUTER APPLY (
        SELECT TOP 1 so2.status
        FROM tbl_trans_sales_opportunity so2
        WHERE so2.document_no = sq.document_no
        ORDER BY so2.id DESC
    ) so
-- §3.2/§6.3 (Phase 4 item 4.1): entry now keys off the explicit REQUEST FOR ENGR.
-- grant (is_requested_for_engr), not "happens to have a project name and any
-- wiring row" - that implicit condition fired with no deliberate sales action and
-- could never distinguish which quotations were actually sent to engineering.
WHERE sq.is_requested_for_engr = 1
