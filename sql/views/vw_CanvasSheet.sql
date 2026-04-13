ALTER VIEW [dbo].[vw_CanvasSheet] AS
SELECT id,
    start_date,
    validity AS TotalDays,
    DATEADD(DAY, validity, start_date) AS ExpirationDate,
    DATEDIFF(
        DAY,
        GETDATE(),
        DATEADD(DAY, validity, start_date)
    ) AS RemainingDays,
    CASE
        WHEN DATEDIFF(
            DAY,
            GETDATE(),
            DATEADD(DAY, validity, start_date)
        ) > 0 THEN 'Active'
        ELSE 'Expired'
    END AS Status
FROM SalesCanvasSheet;