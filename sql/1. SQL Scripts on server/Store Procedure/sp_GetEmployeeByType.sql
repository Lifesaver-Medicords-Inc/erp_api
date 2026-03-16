CREATE PROCEDURE [dbo].[sp_GetEmployeeByType] 
	-- Add the parameters for the stored procedure here
	@EmployeeId VARCHAR(100) 
	
AS
BEGIN
	SET NOCOUNT ON;

    -- Insert statements for procedure here
	SELECT 
		*
	FROM tbl_setup_users WHERE department = @EmployeeId
END
GO


