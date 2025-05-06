package models

import "time"

// Type 活动类型结构体
type Type struct {
	ID   int64  `json:"id" gorm:"column:type_id"`
	Name string `json:"name" gorm:"column:type_name"`
}

// TypeDetail 活动类型详情的接口结构体
type TypeDetail struct {
	ID           int64     `json:"id" gorm:"column:type_id"`
	Name         string    `json:"name" gorm:"column:type_name"`
	Introduction string    `json:"introduction,omitempty" gorm:"column:introduction"`
	CreateTime   time.Time `json:"create_time" gorm:"column:create_time"`
}
