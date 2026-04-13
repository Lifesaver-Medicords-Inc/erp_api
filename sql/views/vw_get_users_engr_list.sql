ALTER VIEW [dbo].[vw_get_users_engr_list] AS
SELECT u.id,
    u.first_name,
    u.last_name,
    u.first_name + ' ' + u.last_name AS full_name,
    u.department
FROM tbl_setup_users u
WHERE department = 'Engineering'