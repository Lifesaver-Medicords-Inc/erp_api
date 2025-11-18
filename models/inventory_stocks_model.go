package models

type InventoryStocksContent struct {
	ReceivingReportId        uint   `json:"receiving_report_id"`
	ReceivingReportDetailsId uint   `json:"receiving_report_details_id"`
	PurchaseOrderDetailsId   uint   `json:"purchase_order_details_id"`
	ReceivingReportDoc       string `json:"receiving_report_doc"`
	PurchaseOrderDoc         string `json:"purchase_order_doc"`
	ItemId                   uint   `json:"item_id"`
	ItemCode                 string `json:"item_code"`
	BinLocation              string `json:"bin_location"`
	QtyIn                    uint   `json:"qty_in"`
	QtyOut                   uint   `json:"qty_out"`
	Uom                      string `json:"uom"`
	SupplierName             string `json:"supplier_name"`
	DateReceived             string `json:"date_received"`
	WarehouseName            string `json:"warehouse_name"`
	WarehouseId              uint   `json:"warehouse_id"`
}

type InventoryStocks struct {
	ID uint `gorm:"primarykey" json:"id"`
	InventoryStocksContent
}

func (InventoryStocks) TableName() string {
	return "tbl_inv_stocks_location"
}

type InventoryStocksAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	InventoryStocksContent
	At
}

func (InventoryStocksAt) TableName() string {
	return "z_tbl_inv_stocks_location_at"
}

type InventoryStocksHistoryContent struct {
	ItemRequestId        uint   `json:"item_request_id"`
	ItemRequestDetailsId uint   `json:"item_request_details_id"`
	InventoryStockId     uint   `json:"inventory_stock_id"`
	ItemId               uint   `json:"item_id"`
	BinLocation          string `json:"bin_location"`
	StockQty             uint   `json:"stock_qty"`
	ReqQty               uint   `json:"req_qty"`
	TransactionDate      string `json:"transaction_date"`
}

type InventoryStocksHistory struct {
	ID uint `gorm:"primarykey" json:"id"`
	InventoryStocksHistoryContent
}

func (InventoryStocksHistory) TableName() string {
	return "tbl_inv_stocks_location_history"
}

type InventoryStocksHistoryAt struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	RefId uint   `json:"ref_id"`
	Code  string `json:"code"`
	InventoryStocksHistoryContent
	At
}

func (InventoryStocksHistoryAt) TableName() string {
	return "z_tbl_inv_stocks_location_history_at"
}
