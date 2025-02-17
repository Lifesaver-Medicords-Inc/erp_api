package models

type ProjectContent struct {
	ProjectName  string `json:"project_name"`
	CustomerName string `json:"customer_name"`
	WebSocketId  uint   `json:"web_socket_id"`
}

type Project struct {
	ID uint `gorm:"primarykey" json:"id"`
	ProjectContent
}

func (Project) TableName() string {
	return "tbl_project_testing"
}
