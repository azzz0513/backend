package models

import "time"

// Checkin 打卡活动类型
type Checkin struct {
	// 基础字段
	ID       int64  `json:"id,string" gorm:"column:checkin_id"`
	AuthorID int64  `json:"author_id" gorm:"column:author_id"`
	TypeID   int64  `json:"type_id" gorm:"column:type_id" binding:"required"`
	WayID    int64  `json:"way_id" gorm:"column:way_id" binding:"required"`
	ListID   int64  `json:"list_id,string" gorm:"column:list_id" binding:"required"`
	Status   int32  `json:"status" gorm:"column:status"`
	Title    string `json:"title" gorm:"column:title" binding:"required"`
	Content  string `json:"content" gorm:"column:content" binding:"required"`

	// 时间字段
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"`

	// 验证码打卡字段
	Password string `json:"password,omitempty" gorm:"column:password"`

	// 长期考勤字段
	StartDate     time.Time `json:"start_date,omitempty" gorm:"column:start_date"`
	EndDate       time.Time `json:"end_date,omitempty" gorm:"column:end_date"`
	DailyDeadline string    `json:"daily_deadline,omitempty" gorm:"column:daily_deadline"`

	// 一次性签到字段
	StartTime       time.Time `json:"start_time,omitempty" gorm:"column:start_time"`
	DurationMinutes uint      `json:"duration_minutes,omitempty" gorm:"column:duration_minutes"`
}

// CheckinDetail 打卡活动详情
type CheckinDetail struct {
	Count   int           `json:"count"`
	Members []*UserDetail `json:"members"`
	*Checkin
}

// CheckinRecord 打卡记录接口结构体
type CheckinRecord struct {
	CheckinID int64     `json:"checkin_id" gorm:"column:checkin_id"`
	UserID    int64     `json:"user_id" gorm:"column:user_id"`
	ListID    int64     `json:"list_id" gorm:"column:list_id"`
	IsChecked int8      `json:"is_checked" gorm:"column:is_checked"`
	CheckTime time.Time `json:"check_time" gorm:"column:check_time"`
}

// CheckinMsg 打卡活动基础信息结构体
type CheckinMsg struct {
	CheckinID  int64     `json:"checkin_id"`
	Type       string    `json:"type"`
	Way        string    `json:"way"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// MsgParticipant 活动参与者获取的打卡活动信息结构体
type MsgParticipant struct {
	AuthorName string `json:"author_name"`
	*CheckinMsg
}

// MsgCreator 活动创建者获取的打卡活动信息结构体
type MsgCreator struct {
	ListName string `json:"list_name"`
	*CheckinMsg
}

// MsgHistory 参与活动历史记录信息结构体
type MsgHistory struct {
	CheckTime time.Time `json:"check_time"`
	*MsgParticipant
}
