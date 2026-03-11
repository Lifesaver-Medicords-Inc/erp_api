CREATE
OR ALTER VIEW [dbo].[vw_get_coa_setup] AS
SELECT a.id AS id,
    a.name,
    b.type
FROM tbl_setup_chart_of_accounts a
    LEFT JOIN tbl_setup_chart_class b ON a.class_id = b.id