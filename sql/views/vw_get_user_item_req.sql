IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_user_item_req' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_user_item_req] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_user_item_req] AS
SELECT a.id AS user_id,
    ISNULL(a.first_name, '') + ' ' + ISNULL(a.last_name, '') AS user_name
FROM tbl_setup_users AS a;