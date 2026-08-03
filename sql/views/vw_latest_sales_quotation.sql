ALTER VIEW [dbo].[vw_latest_sales_quotation] AS WITH LatestQuotes AS (
    SELECT *,
        ROW_NUMBER() OVER (
            PARTITION BY document_no
            -- version_no/sub_version_no are VARCHAR, so ordering by them
            -- directly sorts lexicographically ("9" > "10"), picking a
            -- stale version as "latest" once a document passes 9
            -- revisions. TRY_CAST to numeric for a correct sort, and add
            -- id DESC as a deterministic tiebreaker for rows that would
            -- otherwise tie (e.g. duplicate version numbers).
            ORDER BY TRY_CAST(version_no AS INT) DESC,
                TRY_CAST(sub_version_no AS INT) DESC,
                id DESC
        ) AS rn
    FROM dbo.tbl_trans_sales_quotation
)
SELECT *
FROM LatestQuotes
WHERE rn = 1;