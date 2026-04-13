ALTER VIEW [dbo].[GetBpiCustomer] AS
SELECT p.id AS bpi_id,
    g.branch_name,
    g.customer_code
FROM tbl_bpi p
    INNER JOIN tbl_bpi_general g ON p.id = g.based_id;