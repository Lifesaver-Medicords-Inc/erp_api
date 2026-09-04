package accounting_models

// CreditMemoCustomerView backs the Customer Credit Memo screen's own partner
// picker. Separate from CustomerView (vw_get_customer, Sales Invoice's picker)
// because the two identify a customer by different ids on purpose:
//
//	CustomerView.CustomerID           = tbl_bpi.id        (parent company)
//	CreditMemoCustomerView.PartnerID  = tbl_bpi_general.id (branch)
//
// Credit Memo's own guard (partnerHasEntityType) checks tbl_bpi_entity, which
// keys on bpi_general_id - so the parent id can never satisfy it, and feeding
// it one failed every customer Credit Memo with "partner <n> is not registered
// as a Customer". See vw_get_credit_memo_customer.sql for the full note.
type CreditMemoCustomerView struct {
	PartnerID       uint   `json:"partner_id"`
	ParentBpiID     uint   `json:"parent_bpi_id"`
	Customer        string `json:"customer"`
	CustomerCode    string `json:"customer_code"`
	PaymentTerm     string `json:"payment_term"`
	TaxCode         string `json:"tax_code"`
	CustomerAddress string `json:"customer_address"`
	Tin             string `json:"tin"`
}

func (CreditMemoCustomerView) TableName() string {
	return "vw_get_credit_memo_customer"
}
