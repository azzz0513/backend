package models

import "time"

// MemberList 定义成员列表的结构体
type MemberList struct {
	ID          int64     `json:"id,string" gorm:"column:list_id"`                      // 列表id，无需填写
	AuthorID    int64     `json:"author_id" gorm:"column:author_id"`                    // 作者id，无需填写
	MemberCount int64     `json:"member_count" gorm:"column:member_count"`              // 列表成员数，无需填写
	ListName    string    `json:"list_name" gorm:"column:list_name" binding:"required"` // 列表名称，必填字段
	CreateTime  time.Time `json:"create_time" gorm:"column:create_time"`                // 创建时间，无需填写
}

// UpdateMember 定义修改用户列表的参数结构体
type UpdateMember struct {
	ListID   int64 `json:"list_id,string" gorm:"column:list_id" binding:"required"`   // 列表id，必填字段
	MemberID int64 `json:"member_id,string" gorm:"column:user_id" binding:"required"` // 成员id，必填字段
}

// ListDetail 定义列表详情的参数结构体
type ListDetail struct {
	MemberID   string `json:"member_id" gorm:"column:user_id"`    // 用户id
	MemberName string `json:"member_name" gorm:"column:username"` // 用户名
}
