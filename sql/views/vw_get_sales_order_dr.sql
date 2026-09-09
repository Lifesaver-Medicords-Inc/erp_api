IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_sales_order_dr' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_sales_order_dr] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_sales_order_dr] AS
SELECT a.order_id,
    a.document_no,
    a.project_name,
    a.total_amount_due AS total_amount,
    c.name AS company_name,
    a.sales_executive
FROM tbl_trans_sales_order a
    LEFT JOIN tbl_bpi c ON a.customer_id = c.id