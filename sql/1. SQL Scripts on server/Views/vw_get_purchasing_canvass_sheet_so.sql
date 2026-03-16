CREATE VIEW [dbo].[vw_get_purchasing_canvass_sheet_so] AS
SELECT 
    id,
    supplier_id,
    supplier_name,
    contact_no,
    order_size,
	supplier_stock,
    previous_list_price,
    current_list_price,    
    new_list_price,
    discount,
    net_price,
    payment_terms,
    lead_time,
    item_id,
    start_date,
    CASE 
        WHEN DATEDIFF(DAY, CAST(GETDATE() AS DATE), DATEADD(DAY, price_validity, CAST(start_date AS DATE))) >= 0 
        THEN DATEDIFF(DAY, CAST(GETDATE() AS DATE), DATEADD(DAY, price_validity, CAST(start_date AS DATE)))
        ELSE 0
    END AS price_validity,
    CASE 
        WHEN DATEADD(DAY, price_validity, CAST(start_date AS DATE)) >= CAST(GETDATE() AS DATE) THEN 'Active'
        ELSE 'Expired'
    END AS status,
    CASE
        WHEN current_list_price > previous_list_price THEN N'▲'
        WHEN current_list_price < previous_list_price THEN N'▼'
        ELSE '~'
    END AS price_trend
FROM tbl_purchasing_canvass_sheet_so;

GO


