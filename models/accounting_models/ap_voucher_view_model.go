package accounting_models

type APVoucherPaymentView struct {
	APVoucherId       int     `json:"ap_voucher_id"`
	Supplier          string  `json:"supplier"`
	SupplierCode      string  `json:"supplier_code"`
	Currency          string  `json:"currency"`
	DocNo             string  `json:"doc_no"`
	DocDate           string  `json:"doc_date"`
	TransactionAmount float64 `json:"transaction_amount"`
}

type APVoucherPaymentDetailsView struct {
	APVoucherDetailsId int     `json:"ap_voucher_details_id"`
	APVoucherId        int     `json:"ap_voucher_id"`
	DocNo              string  `json:"doc_no"`
	DueDate            string  `json:"due_date"`
	TransAmount        float64 `json:"trans_amount"`
}
