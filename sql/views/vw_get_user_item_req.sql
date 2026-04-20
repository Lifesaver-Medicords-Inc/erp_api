ALTER VIEW [dbo].[vw_get_user_item_req] AS
SELECT a.id AS user_id,
    ISNULL(a.first_name, '') + ' ' + ISNULL(a.last_name, '') AS user_name
FROM tbl_setup_users AS a;