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
	UserID   int64  `json:"user_id" gorm:"column:user_id"`     // 用户id，无需填写
	UserName string `json:"user_name" gorm:"column:user_name"` // 用户名，需要修改可以填写
	Email    string `json:"email" gorm:"column:email"`         // 邮件地址，需要修改可以填写
}
