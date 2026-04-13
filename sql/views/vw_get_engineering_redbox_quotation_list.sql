ALTER VIEW [dbo].[vw_get_engineering_redbox_quotation_list] AS
SELECT sq.id AS id,
    bpi.name AS client_name,
    sq.document_no AS sales_quotation,
    so.status AS status,
    sq.project_name AS project_name,
    sq.created_by AS sales_executive,
    sq.purpose AS remark
FROM tbl_trans_sales_quotation sq
    INNER JOIN tbl_bpi bpi ON sq.customer_id = bpi.id
    INNER JOIN tbl_trans_sales_opportunity so ON sq.document_no = so.document_no
WHERE sq.project_name IS NOT NULL
    AND LTRIM(RTRIM(sq.project_name)) <> '';