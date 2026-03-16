CREATE VIEW [dbo].[vw_get_chart_of_account] AS 

SELECT 
	a.*,
	b.name AS class_name
FROM tbl_setup_chart_of_accounts a
LEFT JOIN tbl_setup_chart_class b  ON a.class_id = b.id


GO
