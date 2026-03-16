CREATE VIEW [dbo].[vw_get_all_user_list] AS
SELECT
    ISNULL(a.first_name, '') + ' ' + ISNULL(a.last_name, '') AS user_name
FROM tbl_setup_users AS a;

GO

