package logic

import (
	"fmt"
	"go.uber.org/zap"
	"web_app/dao/mysql"
	"web_app/models"
	"web_app/pkg/snowflake"
)

// CreateMemberList 创建成员列表
func CreateMemberList(l *models.MemberList) (err error) {
	// 1.生成list id
	l.ID = snowflake.GenID()
	// 2.保存到数据库
	err = mysql.CreateMemberList(l)
	if err != nil {
		zap.L().Error("mysql.CreateMemberList failed", zap.Error(err))
		return
	}
	// 3.返回
	return
}

// AddMember 添加成员
func AddMember(m *models.UpdateMember) (err error) {
	zap.L().Debug("AddMember",
		zap.Int64("list_id", m.ListID),
		zap.Int64("member_id", m.MemberID))
	return mysql.AddMember(m)
}

// DeleteMember 删除成员
func DeleteMember(m *models.UpdateMember) (err error) {
	zap.L().Debug("DeleteMember",
		zap.Int64("list_id", m.ListID),
		zap.Int64("member_id", m.MemberID))
	return mysql.DeleteMember(m)
}

// GetListList 获取当前用户创建的成员列表
func GetListList(authorID, page, size int64) (data []*models.MemberList, err error) {
	// 根据创建者的ID取得用户列表
	data, err = mysql.GetListList(authorID, page, size)
	if err != nil {
		zap.L().Error("mysql.GetListList failed", zap.Error(err))
		return
	}
	return
}

// GetListDetail 获取成员列表详情
func GetListDetail(pid, page, size int64) (data []*models.ListDetail, err error) {
	// 根据列表id取得活动列表
	data, err = mysql.GetListDetail(pid, page, size)
	if err != nil {
		zap.L().Error("mysql.GetListDetail failed", zap.Error(err))
		return
	}
	return
}

// GetJoinURL 获取参与成员列表的id
func GetJoinURL(checkinID int64) (url string, err error) {
	url = fmt.Sprintf("http://3.138.230.142:8087/join-list.html/%d", checkinID)
	return
}
