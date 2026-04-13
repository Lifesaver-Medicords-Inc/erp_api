ALTER VIEW [dbo].[vw_get_purchasing_guiding_price] AS WITH Prices AS (
    SELECT b.item_id,
        b.discounted_price,
        a.date,
        a.supplier_name
    FROM tbl_purchasing_purchase_order a
        LEFT JOIN tbl_purchasing_purchase_order_details b ON a.id = b.based_id
),
-- Remove consecutive duplicate prices
DistinctPrices AS (
    SELECT *,
        LAG(discounted_price) OVER (
            PARTITION BY item_id
            ORDER BY date DESC
        ) AS prev_price
    FROM Prices
),
FilteredPrices AS (
    SELECT item_id,
        discounted_price,
        date,
        supplier_name
    FROM DistinctPrices
    WHERE prev_price IS NULL
        OR prev_price <> discounted_price
),
-- Rank latest distinct prices
RankedDistinct AS (
    SELECT *,
        ROW_NUMBER() OVER (
            PARTITION BY item_id
            ORDER BY date DESC
        ) AS rn
    FROM FilteredPrices
),
-- Get first date per item for cutoff checks
PriceDateRanges AS (
    SELECT item_id,
        MIN(date) AS first_purchase_date
    FROM Prices
    GROUP BY item_id
),
-- Calculate lowest prices and corresponding supplier_name
LowestPrices AS (
    SELECT d.item_id,
        -- Lowest 1 Year
        CASE
            WHEN DATEDIFF(DAY, d.first_purchase_date, GETDATE()) >= 365 THEN CAST(lp1.discounted_price AS VARCHAR)
            ELSE '-'
        END AS lowest_1yr,
        CASE
            WHEN DATEDIFF(DAY, d.first_purchase_date, GETDATE()) >= 365 THEN ISNULL(lp1.supplier_name, '-')
            ELSE '-'
        END AS lowest_1yr_supplier_name,
        -- Lowest 3 Year
        CASE
            WHEN DATEDIFF(DAY, d.first_purchase_date, GETDATE()) >= 3 * 365 THEN CAST(lp3.discounted_price AS VARCHAR)
            ELSE '-'
        END AS lowest_3yr,
        CASE
            WHEN DATEDIFF(DAY, d.first_purchase_date, GETDATE()) >= 3 * 365 THEN ISNULL(lp3.supplier_name, '-')
            ELSE '-'
        END AS lowest_3yr_supplier_name,
        -- Lowest All Time
        CAST(lpa.discounted_price AS VARCHAR) AS lowest_alltime,
        ISNULL(lpa.supplier_name, '-') AS lowest_alltime_supplier_name
    FROM PriceDateRanges d -- Lowest 1yr
        LEFT JOIN (
            SELECT p1.item_id,
                p1.discounted_price,
                p1.supplier_name
            FROM Prices p1
                JOIN (
                    SELECT item_id,
                        MIN(discounted_price) AS min_price
                    FROM Prices
                    WHERE date >= DATEADD(YEAR, -1, GETDATE())
                    GROUP BY item_id
                ) x ON p1.item_id = x.item_id
                AND p1.discounted_price = x.min_price
            WHERE p1.date >= DATEADD(YEAR, -1, GETDATE())
        ) lp1 ON d.item_id = lp1.item_id -- Lowest 3yr
        LEFT JOIN (
            SELECT p1.item_id,
                p1.discounted_price,
                p1.supplier_name
            FROM Prices p1
                JOIN (
                    SELECT item_id,
                        MIN(discounted_price) AS min_price
                    FROM Prices
                    WHERE date >= DATEADD(YEAR, -3, GETDATE())
                    GROUP BY item_id
                ) x ON p1.item_id = x.item_id
                AND p1.discounted_price = x.min_price
            WHERE p1.date >= DATEADD(YEAR, -3, GETDATE())
        ) lp3 ON d.item_id = lp3.item_id -- Lowest All Time
        LEFT JOIN (
            SELECT p1.item_id,
                p1.discounted_price,
                p1.supplier_name
            FROM Prices p1
                JOIN (
                    SELECT item_id,
                        MIN(discounted_price) AS min_price
                    FROM Prices
                    GROUP BY item_id
                ) x ON p1.item_id = x.item_id
                AND p1.discounted_price = x.min_price
        ) lpa ON d.item_id = lpa.item_id
)
SELECT r.item_id,
    -- LAST
    ISNULL(
        CAST(
            MAX(
                CASE
                    WHEN r.rn = 1 THEN r.discounted_price
                END
            ) AS VARCHAR
        ),
        0
    ) AS last_price,
    ISNULL(
        MAX(
            CASE
                WHEN r.rn = 1 THEN r.supplier_name
            END
        ),
        '-'
    ) AS last_supplier_name,
    -- 2ND
    ISNULL(
        CAST(
            MAX(
                CASE
                    WHEN r.rn = 2 THEN r.discounted_price
                END
            ) AS VARCHAR
        ),
        0
    ) AS second_last_price,
    ISNULL(
        MAX(
            CASE
                WHEN r.rn = 2 THEN r.supplier_name
            END
        ),
        '-'
    ) AS second_last_supplier_name,
    -- 3RD
    ISNULL(
        CAST(
            MAX(
                CASE
                    WHEN r.rn = 3 THEN r.discounted_price
                END
            ) AS VARCHAR
        ),
        0
    ) AS third_last_price,
    ISNULL(
        MAX(
            CASE
                WHEN r.rn = 3 THEN r.supplier_name
            END
        ),
        '-'
    ) AS third_last_supplier_name,
    -- LOWEST PRICES
    lp.lowest_1yr,
    lp.lowest_1yr_supplier_name,
    lp.lowest_3yr,
    lp.lowest_3yr_supplier_name,
    lp.lowest_alltime,
    lp.lowest_alltime_supplier_name
FROM RankedDistinct r
    JOIN LowestPrices lp ON r.item_id = lp.item_id
GROUP BY r.item_id,
    lp.lowest_1yr,
    lp.lowest_1yr_supplier_name,
    lp.lowest_3yr,
    lp.lowest_3yr_supplier_name,
    lp.lowest_alltime,
    lp.lowest_alltime_supplier_name;