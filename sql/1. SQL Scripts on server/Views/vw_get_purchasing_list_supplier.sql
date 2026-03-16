CREATE VIEW [dbo].[vw_get_purchasing_list_supplier] AS

SELECT 
    b.id AS supplier_id,
	MAX(b.branch_tel_no) AS main_tel_no,
    MAX(b.branch_name) AS supplier_name,
    MAX(b.supplier_code) AS supplier_code,
    MAX(a.tin) AS tin_no,
    MAX(b.fax_no) AS fax_no,
    MAX(h.tax_code) AS tax_code,
	MAX(h.payment_terms_id) AS payment_terms_id,
    MAX(addresses.address) AS address,
	MAX(items.item_ids) AS item_ids
FROM tbl_bpi a
LEFT JOIN tbl_bpi_general b ON a.id = b.based_id
LEFT JOIN tbl_bpi_entity d ON b.id = d.bpi_general_id
LEFT JOIN tbl_setup_bpi_entity e ON d.entity_id = e.id
LEFT JOIN tbl_bpi_industries f ON b.based_id = f.bpi_id
LEFT JOIN tbl_bpi_items h ON b.id = h.branch_id
LEFT JOIN tbl_bpi_address i ON b.id = i.branch_id

-- OUTER APPLY for distinct addresses
OUTER APPLY (
    SELECT 
        STUFF((
            SELECT DISTINCT ', ' + CAST(a2.location AS NVARCHAR(MAX))
            FROM tbl_bpi_address a2
            WHERE a2.branch_id = b.id AND a2.location IS NOT NULL
            FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'), 1, 2, '') AS address
) addresses

-- OUTER APPLY for item IDs
OUTER APPLY (
    SELECT 
        STUFF((
            SELECT DISTINCT ', ' + CAST(h2.item_id AS NVARCHAR(MAX))
            FROM tbl_bpi_items h2
            WHERE h2.branch_id = b.id AND h2.item_id IS NOT NULL
            FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'), 1, 2, '') AS item_ids
) items

WHERE e.code = 'SUPPLIER'

GROUP BY b.id;

GO


