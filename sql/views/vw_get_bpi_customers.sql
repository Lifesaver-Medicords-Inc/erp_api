ALTER VIEW [dbo].[vw_get_bpi_customers] AS
SELECT a.id AS bpi_id,
    b.branch_name,
    b.customer_code
FROM tbl_bpi a
    INNER JOIN tbl_bpi_general b ON a.id = b.based_id
-- Fix: tbl_setup_bpi_entity.code for Customer is 'CUS' (SMPC_ERP_SPEC §17.3's
-- authoritative Code|Name table), not 'CUSTOMER' - this view's own filter
-- never matched a real code, so it silently returned zero rows for everyone.
-- Found while implementing the same Customer-entity filter for Bug #018
-- (vw_get_CRM) and Bug #024 (Sales Quotation's customer picker).
WHERE EXISTS (
        SELECT 1
        FROM tbl_bpi_entity c
            INNER JOIN tbl_setup_bpi_entity d ON c.entity_id = d.id
        WHERE c.bpi_general_id = b.id
            AND d.code = 'CUS'
    );