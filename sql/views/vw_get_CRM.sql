ALTER VIEW [dbo].[vw_get_CRM] AS WITH LatestCRM AS (
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
    -- Bug #018 (Trello, "Supplier Entity Type can be seen in the CRM"): this
    -- view had no entity-type filter at all, so every BPI branch showed up in
    -- every sales exec's CRM regardless of type. Spec (SMPC_ERP_SPEC §4.1.5/
    -- outbound interfaces table) is explicit that CRM is Customer-entities
    -- only. Same EXISTS-against-tbl_bpi_entity pattern already used by
    -- vw_get_bpi_customers.sql.
    WHERE EXISTS (
        SELECT 1
        FROM dbo.tbl_bpi_entity AS e
            INNER JOIN dbo.tbl_setup_bpi_entity AS f ON e.entity_id = f.id
        WHERE e.bpi_general_id = a.id
            AND f.code = 'CUSTOMER'
    )
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