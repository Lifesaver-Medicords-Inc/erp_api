ALTER PROCEDURE [dbo].[sp_chart_of_account_classification] -- Add the parameters for the stored procedure here
@code nvarchar(50) = null AS BEGIN -- SET NOCOUNT ON added to prevent extra result sets from
-- interfering with SELECT statements.
SET NOCOUNT ON;
SELECT a.id,
    a.code,
    a.name
FROM tbl_setup_chart_of_accounts a
    LEFT JOIN tbl_setup_chart_class b ON a.class_id = b.id
WHERE b.code = @code;
END;