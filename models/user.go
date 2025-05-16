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

// ChangePassword 修改用户密码的结构体
type ChangePassword struct {
	UserID      int64  `json:"user_id" gorm:"column:user_id"`                                              // 用户id，无需填写
	OldPassword string `json:"old_password" binding:"required"`                                            // 用户旧密码，必填字段
	NewPassword string `json:"new_password" gorm:"column:password" binding:"required,nefield=OldPassword"` // 用户新密码（与旧密码必须不同），必填字段
	RePassword  string `json:"re_password" binding:"required,eqfield=NewPassword"`                         // 用户新密码确认，必填字段
}
