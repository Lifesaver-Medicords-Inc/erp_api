IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_users_engr_list' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_users_engr_list] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_users_engr_list] AS
SELECT u.id,
    u.first_name,
    u.last_name,
    u.first_name + ' ' + u.last_name AS full_name,
    u.department
FROM tbl_setup_users u
    LEFT JOIN tbl_position p ON u.position_id = p.id
WHERE UPPER(LTRIM(RTRIM(ISNULL(u.department, '')))) = 'ENGINEERING'
    OR UPPER(LTRIM(RTRIM(ISNULL(p.name, '')))) = 'ENGINEER'