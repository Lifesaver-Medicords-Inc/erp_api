CREATE
OR ALTER VIEW [dbo].[vw_get_CRM] AS WITH LatestCRM AS (
    SELECT ISNULL(a.id, 0) AS id,
        a.branch_name,
        b.number,
        b.name,
        b.email,
        c.tag,
        c.date,
        c.remark,
        d.code,
        a.sales_id,
        ISNULL(c.crm_id, 0) AS crm_id,
        ROW_NUMBER() OVER (
            PARTITION BY a.id
            ORDER BY c.date DESC
        ) AS rn -- Rank by date in descending order
    FROM dbo.tbl_bpi_general AS a
        LEFT OUTER JOIN dbo.tbl_bpi_contacts AS b ON a.based_id = b.based_id
        LEFT OUTER JOIN dbo.tbl_trans_sales_crm AS c ON a.id = c.based_id
        LEFT OUTER JOIN dbo.tbl_position AS d ON b.position = d.id
)
SELECT id,
    branch_name,
    number,
    CONCAT(name, ' - ', code) AS name,
    email,
    tag,
    date,
    remark,
    crm_id,
    sales_id
FROM LatestCRM
WHERE rn = 1 -- Get only the latest entry for each `id`