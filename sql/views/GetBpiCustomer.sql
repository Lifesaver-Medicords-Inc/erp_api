IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'GetBpiCustomer' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[GetBpiCustomer] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[GetBpiCustomer] AS
SELECT p.id AS bpi_id,
    g.branch_name,
    g.customer_code
FROM tbl_bpi p
    INNER JOIN tbl_bpi_general g ON p.id = g.based_id;