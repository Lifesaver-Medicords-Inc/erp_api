package models
type UserContent struct {
    FirstName  string `json:"first_name"`
    LastName   string `json:"last_name"`
    Department string `json:"department"`
    Position   string `json:"position"`
    EmployeeId string `gorm:"unique" json:"employee_id"`
    Password   string `json:"password"`
}
type User struct {
    ID uint `gorm:"primarykey" json:"id"`
    UserContent
}
func (User) TableName() string {
    return "tbl_setup_users"
}
type UserAt struct {
    ID    uint `gorm:"primarykey" json:"id"`
    RefId uint `json:"ref_id"`
    UserContent
    At
}
func (UserAt) TableName() string {
    return "z_tbl_setup_users_at"
}