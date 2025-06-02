package models

type PurchasingRedboxPurchaseListContent struct {
	Id             int    `json:"id"`
	DocNo          string `json:"doc_no"`
	ProjectName    string `json:"project_name"`
	CommitmentDate string `json:"commitment_date"`
	Purchaser      string `json:"purchaser"`
	DetailIds      string `json:"detail_ids"`
	ItemIds        string `json:"item_ids"`
	ItemNames      string `json:"item_names"`
	Customer       string `json:"customer"`
	OrderType      string `json:"order_type"`
}

type PurchasingRedboxPurchaseListView struct {
	PurchasingRedboxPurchaseListContent
}

func (PurchasingRedboxPurchaseListView) TableName() string {
	return "vw_get_purchasing_redbox_purchase_list"
}
