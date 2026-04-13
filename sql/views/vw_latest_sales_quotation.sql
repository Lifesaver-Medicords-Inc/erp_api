ALTER VIEW [dbo].[vw_latest_sales_quotation] AS WITH LatestQuotes AS (
    SELECT *,
        ROW_NUMBER() OVER (
            PARTITION BY document_no
            ORDER BY version_no DESC,
                sub_version_no DESC
        ) AS rn
    FROM dbo.tbl_trans_sales_quotation
)
SELECT *
FROM LatestQuotes
WHERE rn = 1;