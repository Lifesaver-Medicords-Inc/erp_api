IF NOT EXISTS (SELECT 1 FROM sys.procedures WHERE name = 'sp_GetEmployeeByType' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE PROCEDURE [dbo].[sp_GetEmployeeByType] AS SET NOCOUNT ON;')
END
GO
ALTER PROCEDURE [dbo].[sp_GetEmployeeByType] -- Add the parameters for the stored procedure here
@EmployeeId VARCHAR(100) AS BEGIN
SET NOCOUNT ON;
-- Insert statements for procedure here
SELECT *
FROM tbl_setup_users
WHERE department = @EmployeeId
END