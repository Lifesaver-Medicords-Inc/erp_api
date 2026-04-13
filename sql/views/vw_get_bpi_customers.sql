ALTER VIEW [dbo].[vw_get_bpi_customers] AS
SELECT a.id AS bpi_id,
    b.branch_name,
    b.customer_code
FROM tbl_bpi a
    INNER JOIN tbl_bpi_general b ON a.id = b.based_id
WHERE EXISTS (
        SELECT 1
        FROM tbl_bpi_entity c
            INNER JOIN tbl_setup_bpi_entity d ON c.entity_id = d.id
        WHERE c.bpi_general_id = b.id
            AND d.code = 'CUSTOMER'
    );