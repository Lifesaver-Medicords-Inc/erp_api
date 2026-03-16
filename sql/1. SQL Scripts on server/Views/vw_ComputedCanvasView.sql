CREATE VIEW [dbo].[vw_ComputedCanvasView] AS
SELECT 
    id,
    net_price,
    discount,
    unit_price,
    validity,
    lead_time,
    item_based_id,       -- Add this
    supplier_based_id,   -- Add this
    CASE 
        WHEN ISDATE(validity) = 1 THEN CAST(validity AS DATE)  
        ELSE DATEADD(DAY, CAST(validity AS INT), GETDATE())  
    END AS ExpirationDate,
    
    DATEDIFF(DAY, GETDATE(), 
        CASE 
            WHEN ISDATE(validity) = 1 THEN CAST(validity AS DATE)
            ELSE DATEADD(DAY, CAST(validity AS INT), GETDATE())
        END
    ) AS RemainingDays
FROM tbl_sales_canvas_sheet
WHERE net_price > 0;
GO


