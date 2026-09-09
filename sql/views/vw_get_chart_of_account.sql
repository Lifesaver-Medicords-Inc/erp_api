IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_chart_of_account' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_chart_of_account] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_chart_of_account] AS
SELECT a.*,
    b.name AS class_name
FROM tbl_setup_chart_of_accounts a
    LEFT JOIN tbl_setup_chart_class b ON a.class_id = b.id