ALTER VIEW [dbo].[vw_get_engineering_redbox_quotation_list] AS
SELECT sq.id AS id,
    bpi.name AS client_name,
    sq.document_no AS sales_quotation,
    so.status AS status,
    sq.project_name AS project_name,
    sq.created_by AS sales_executive,
    sq.purpose AS remark
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
WHERE sq.project_name IS NOT NULL
    AND LTRIM(RTRIM(sq.project_name)) <> ''
    AND EXISTS (
        SELECT 1
        FROM dbo.tbl_trans_sales_project_wiring w
		INNER JOIN tbl_trans_sales_project_item_set spis ON w.based_id = spis.item_set_id
        WHERE spis.based_id = sq.id 
    ) --Checking if the project quotation have wiring (requested by the sales)
