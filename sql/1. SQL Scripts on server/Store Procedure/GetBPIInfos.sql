CREATE PROCEDURE [dbo].[GetBPIInfos]
AS
BEGIN
    SELECT
        a.id,
        a.sales_id,
        a.name,
        a.main_website,
        a.tin,
        a.main_website,
        b.industry_id,
        b.bpi_id,
        c.name AS industry_name
    FROM tbl_bpi a
    LEFT JOIN tbl_bpi_industries b ON a.id = b.bpi_id
    LEFT JOIN tbl_setup_bpi_industries c ON b.industry_id = c.id;
END;
GO


