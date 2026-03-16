CREATE PROCEDURE [dbo].[sp_tax_setup_classification]
	-- Add the parameters for the stored procedure here
		@code nvarchar(20) = null 
AS
BEGIN

	SET NOCOUNT ON;
	SELECT 
		a.id,
		a.code,
		a.input_tax_creditable,
		a.tax_desc,
		b.tax_rate,
		c.id as account_id,
		 'ACTIVE' as status
	FROM tbl_setup_tax a 
	INNER JOIN tbl_setup_tax_details b ON a.id = b.tax_code_id
	INNER JOIN tbl_setup_chart_of_accounts c ON a.coa_sales_id = c.id OR a.coa_purchase_id = c.id
	INNER JOIN tbl_setup_chart_class e ON e.id = c.class_id AND e.code = @code
	WHERE b.valid_from = (
    SELECT MIN(sub.valid_from)
    FROM tbl_setup_tax_details sub
    WHERE sub.tax_code_id = a.id
    AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_tax_details prev
        WHERE prev.tax_code_id = a.id
        AND prev.valid_from < sub.valid_from
        AND (prev.valid_to >= GETDATE() OR prev.valid_to IS NULL)
    )
    AND (sub.valid_to >= GETDATE() OR sub.valid_to IS NULL)
)
END
GO


