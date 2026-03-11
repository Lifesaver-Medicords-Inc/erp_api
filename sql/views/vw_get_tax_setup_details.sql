CREATE
OR ALTER VIEW [dbo].[vw_get_tax_setup_details] AS
SELECT a.id,
    a.tax_code_id,
    a.valid_from,
    a.valid_to,
    a.tax_rate,
    CASE
        -- VALID_TO IS EMPTY ('' or NULL)
        WHEN NULLIF(a.valid_to, '') IS NULL
        AND CAST(GETDATE() AS DATE) >= a.valid_from THEN 'ACTIVE'
        WHEN NULLIF(a.valid_to, '') IS NULL
        AND CAST(GETDATE() AS DATE) < a.valid_from THEN 'INACTIVE' -- VALID_TO HAS VALUE
        WHEN CAST(GETDATE() AS DATE) BETWEEN a.valid_from AND a.valid_to THEN 'ACTIVE'
        WHEN CAST(GETDATE() AS DATE) < a.valid_from THEN 'FUTURE'
        WHEN CAST(GETDATE() AS DATE) > a.valid_to THEN 'INACTIVE'
    END AS valid_status
FROM tbl_setup_tax_details a;