package models

import "time"

// Checkin 打卡活动类型
type Checkin struct {
	// 基础字段
	ID       int64  `json:"id,string" gorm:"column:checkin_id"`                   // 活动id，无需填写
	AuthorID int64  `json:"author_id" gorm:"column:author_id"`                    // 作者id，无需填写
	TypeID   int64  `json:"type_id" gorm:"column:type_id" binding:"required"`     // 活动类型（一次性签到/长期考勤），必填字段
	WayID    int64  `json:"way_id" gorm:"column:way_id" binding:"required"`       // 打卡方式（验证码），必填字段
	ListID   int64  `json:"list_id,string" gorm:"column:list_id"`                 // 用户列表id
	ListName string `json:"list_name" gorm:"column:list_name" binding:"required"` // 用户列表名，必填字段
	Status   int32  `json:"status" gorm:"column:status"`                          // 活动状态，无需填写，默认为有效
	Title    string `json:"title" gorm:"column:title" binding:"required"`         // 活动标题，必填字段
	Content  string `json:"content" gorm:"column:content" binding:"required"`     // 活动内容，必填字段

	// 时间字段
	CreateTime time.Time `json:"create_time" gorm:"column:create_time"` // 活动创建时间，无需填写
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time"` // 活动创建时间，无需填写

	// 验证码打卡字段
	Password string `json:"password,omitempty" gorm:"column:password"` // 验证码，打卡方式为验证码时为必填字段

	// 长期考勤字段
	StartDate     time.Time `json:"start_date,omitempty" gorm:"column:start_date"`         // 活动开始日期，活动为长期考勤类型时为必填字段
	EndDate       time.Time `json:"end_date,omitempty" gorm:"column:end_date"`             // 活动结束日期，活动为长期考勤类型时为必填字段
	DailyDeadline string    `json:"daily_deadline,omitempty" gorm:"column:daily_deadline"` // 每日打卡时限，活动为长期考勤类型时为必填字段

	// 一次性签到字段
	StartTime       time.Time `json:"start_time,omitempty" gorm:"column:start_time"`             // 活动开始时间，活动为一次性签到时为必填字段
	DurationMinutes uint      `json:"duration_minutes,omitempty" gorm:"column:duration_minutes"` // 打卡活动持续时间，分钟数，活动为一次性签到时为必填字段
}

// CheckinDetail 打卡活动详情
type CheckinDetail struct {
	Count    int               `json:"count"`   // 已完成人员数
	Members  []*UserEasyDetail `json:"members"` // 未完成人员列表
	*Checkin                   // 内嵌活动结构体
}

//type CheckinDetail2 struct {
//	CheckedCount int  `json:"checked_count"` // 已完成人数
//	UnCheckedCount int  `json:"un_checked_count"` // 未完成人数
//	CheckedMembers []*UserEasyDetail `json:"checked_members"` // 已完成成员列表
//	UnCheckedMembers []*UserEasyDetail `json:"un_checked_members"` // 未完成成员列表
//	*Checkin // 内嵌活动结构体
//}

// Statistics 长期考勤活动统计数据结构体
type Statistics struct {
	UserID       int64     `json:"user_id" gorm:"column:user_id"`              // 用户id
	DailyCount   int       `json:"daily_count" gorm:"column:checkin_count"`    // 用户日打卡数
	WeeklyCount  int       `json:"weekly_count" gorm:"column:checkin_count"`   // 用户周打卡数
	MonthlyCount int       `json:"monthly_count" gorm:"column:checkin_count"`  // 用户月打卡数
	UserName     string    `json:"user_name" gorm:"column:user_name"`          // 用户名称
	CheckTime    time.Time `json:"check_time" gorm:"column:last_checkin_time"` // 用户最近打卡时间
}

// CheckinRecord 打卡记录接口结构体
type CheckinRecord struct {
	CheckinID int64     `json:"checkin_id" gorm:"column:checkin_id"` // 打卡活动id
	UserID    int64     `json:"user_id" gorm:"column:user_id"`       // 用户id
	ListID    int64     `json:"list_id" gorm:"column:list_id"`       // 用户所属列表id
	IsChecked int8      `json:"is_checked" gorm:"column:is_checked"` // 用户打卡确认
	CheckTime time.Time `json:"check_time" gorm:"column:check_time"` // 打卡时间
}

// CheckinMsg 打卡活动基础信息结构体
type CheckinMsg struct {
	CheckinID  int64     `json:"checkin_id"`  // 打卡活动id
	Type       string    `json:"type"`        // 打卡活动类型
	Way        string    `json:"way"`         // 打卡方式
	Title      string    `json:"title"`       // 打卡活动标题
	Content    string    `json:"content"`     // 打卡活动内容
	CreateTime time.Time `json:"create_time"` // 打卡活动创建时间
	UpdateTime time.Time `json:"update_time"` // 打卡活动更新时间

	// 长期考勤字段
	StartDate     time.Time `json:"start_date" gorm:"column:start_date"`         // 活动开始日期
	EndDate       time.Time `json:"end_date" gorm:"column:end_date"`             // 活动结束日期
	DailyDeadline string    `json:"daily_deadline" gorm:"column:daily_deadline"` // 每日打卡时限

	// 一次性签到字段
	StartTime       time.Time `json:"start_time" gorm:"column:start_time"`             // 活动开始时间
	DurationMinutes uint      `json:"duration_minutes" gorm:"column:duration_minutes"` // 打卡活动持续时间，分钟数
}

// MsgParticipant 活动参与者获取的打卡活动信息结构体
type MsgParticipant struct {
	AuthorName  string `json:"author_name"` // 作者名称
	*CheckinMsg        // 内嵌打卡活动基础信息结构体
}

// MsgCreator 活动创建者获取的打卡活动信息结构体
type MsgCreator struct {
	ListName    string `json:"list_name"` // 用户列表名
	*CheckinMsg        // 内嵌打卡活动基础信息结构体
}

// MsgHistory 参与活动历史记录信息结构体
type MsgHistory struct {
	CheckTime       time.Time `json:"check_time"` // 打卡时间
	*MsgParticipant           // 内嵌活动参与者获取的打卡活动信息结构体
}
