package models

// User 用户结构体
type User struct {
	UserID   int64  `db:"user_id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Token    string
}

// UserDetail 用户详情结构体
type UserDetail struct {
	UserID   int64  `json:"user_id" gorm:"column:user_id"`
	UserName string `json:"user_name" gorm:"column:username"`
}

// UpdateUser 修改用户数据的结构体
type UpdateUser struct {
	UserID   int64  `json:"user_id" gorm:"column:user_id"`
	UserName string `json:"user_name" gorm:"column:user_name"`
	Email    string `json:"email" gorm:"column:email"`
}
