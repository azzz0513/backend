package models

import "time"

// MemberList 定义成员列表的结构体
type MemberList struct {
	ID         int64     `json:"id,string" gorm:"column:list_id"`
	AuthorID   int64     `json:"author_id" gorm:"column:author_id"`
	ListName   string    `json:"list_name" gorm:"column:list_name" binding:"required"`
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
}

// AddMember 定义添加用户到用户列表的参数结构体
type AddMember struct {
	ListID   int64 `json:"list_id,string" gorm:"column:list_id" binding:"required"` // 列表id
	MemberID int64 `json:"member_id,string" gorm:"column:user_id" binding:"required"`
}

// ListDetail 定义列表详情的参数结构体
type ListDetail struct {
	MemberID   string `json:"member_id" gorm:"column:user_id"`
	MemberName string `json:"member_name" gorm:"column:username"`
}
