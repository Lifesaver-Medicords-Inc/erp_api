package accounting_models

// PPECategoryBreakdown — one asset category's rollup as of a given date,
// matching the shape a real PP&E note breaks assets down by (Land,
// Building, Machinery, ...).
type PPECategoryBreakdown struct {
	CategoryId              uint    `json:"category_id"`
	CategoryName            string  `json:"category_name"`
	Cost                    float64 `json:"cost"`
	AccumulatedDepreciation float64 `json:"accumulated_depreciation"`
	NetBookValue            float64 `json:"net_book_value"`
}

// PPEResult — the computed PP&E position as of one date. Nothing here is
// stored; it's derived fresh from tbl_fixed_asset every time, the same
// "compute at read time" approach GetCostOfSales already uses for FIFO
// lot consumption.
type PPEResult struct {
	AsOf                         string                 `json:"as_of"`
	TotalCost                    float64                `json:"total_cost"`
	TotalAccumulatedDepreciation float64                `json:"total_accumulated_depreciation"`
	TotalNetBookValue            float64                `json:"total_net_book_value"`
	Categories                   []PPECategoryBreakdown `json:"categories"`
}
