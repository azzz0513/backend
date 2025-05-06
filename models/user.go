package models

type User struct {
	UserID   int64  `db:"user_id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Token    string
}

type UserDetail struct {
	UserID   int64  `json:"user_id" gorm:"column:user_id"`
	UserName string `json:"user_name" gorm:"column:user_name"`
}
