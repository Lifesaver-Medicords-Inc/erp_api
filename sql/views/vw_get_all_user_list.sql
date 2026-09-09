IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_all_user_list' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_all_user_list] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_all_user_list] AS
SELECT ISNULL(a.first_name, '') + ' ' + ISNULL(a.last_name, '') AS user_name
FROM tbl_setup_users AS a;