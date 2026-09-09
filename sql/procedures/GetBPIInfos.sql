IF NOT EXISTS (SELECT 1 FROM sys.procedures WHERE name = 'GetBPIInfos' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE PROCEDURE [dbo].[GetBPIInfos] AS SET NOCOUNT ON;')
END
GO
ALTER PROCEDURE [dbo].[GetBPIInfos] AS BEGIN
SELECT a.id,
    a.sales_id,
    a.name,
    a.main_website AS website,
    a.tin,
    a.main_tel_no AS tel_no,
    b.industry_id,
    b.bpi_id,
    c.name AS industry_name
FROM tbl_bpi a
    LEFT JOIN tbl_bpi_industries b ON a.id = b.bpi_id
    LEFT JOIN tbl_setup_bpi_industries c ON b.industry_id = c.id;
END;