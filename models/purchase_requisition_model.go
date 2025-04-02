package models

type PurchaseRequisitionContent struct {
	DateRequest   string `json:"date_request"`
	DateRequired  string `json:"date_required"`
	RequestBy     string `json:"request_by"`
	Department    string `json:"department"`
	ContactNo     string `json:"contact_no"`
	DocNo         string `json:"doc_no"`
	Justification string `json:"justification"`
	Remarks       string `json:"remarks"`
	Approval      string `json:"approval"`
	IsApproved    bool   `json:"is_approved"`
	Status        string `json:"status"`
	Purchaser     string `json:"purchaser"`
}

type PurchaseRequisition struct {
	PR_ID uint `gorm:"primarykey" json:"pr_id"`
	PurchaseRequisitionContent
}

func (PurchaseRequisition) TableName() string {
	return "tbl_purchasing_purchase_requisition"
}

type PurchaseRequisitionAt struct {
	ID    uint `gorm:"primarykey" json:"id"`
	RefId uint `json:"ref_id"`
	PurchaseRequisitionContent
	At
}

func (PurchaseRequisitionAt) TableName() string {
	return "z_tbl_purchasing_purchase_requisition_at"
}
