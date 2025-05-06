package mysql

import (
	"go.uber.org/zap"
	"web_app/models"
)

// CreateMemberList 处理创建成员列表的数据库交互
func CreateMemberList(l *models.MemberList) (err error) {
	if err = DB.Table("member_list").Select("list_id", "author_id", "list_name").Create(l).Error; err != nil {
		zap.L().Error("CreateMemberList failed", zap.Error(err))
		return
	}
	return
}

// AddMember 处理往成员列表添加成员的数据库交互
func AddMember(m *models.AddMember) (err error) {
	if err = DB.Table("list_participants").Select("list_id", "user_id").Create(&m).Error; err != nil {
		zap.L().Error("AddMember failed", zap.Error(err))
		return
	}
	return
}

// GetListList 处理获取当前啊用户创建的成员列表的数据库交互
func GetListList(authorID, page, size int64) (data []*models.MemberList, err error) {
	data = make([]*models.MemberList, 0, 2)
	if err = DB.Table("member_list").
		Where("author_id=?", authorID).
		Select("list_id", "author_id", "list_name", "create_time").
		Order("create_time").
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Scan(&data).Error; err != nil {
		zap.L().Error("GetListList failed", zap.Error(err))
		return
	}
	return
}

// GetListDetail 处理获取成员列表详情的数据库交互
func GetListDetail(pid, page, size int64) (data []*models.ListDetail, err error) {
	data = make([]*models.ListDetail, 0, 4)
	if err = DB.Table("list_participants").
		Joins("INNER JOIN users ON list_participants.user_id = users.user_id").
		Where("list_participants.list_id = ?", pid).
		Select("list_participants.user_id", "users.username").
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Scan(&data).Error; err != nil {
		zap.L().Error("GetListDetail failed", zap.Error(err))
		return nil, err
	}
	return
}

// GetListDetailByID 获取指定用户名单的信息
func GetListDetailByID(listID int64) (listName string, err error) {
	if err = DB.Table("member_list").Where("list_id=?", listID).Select("list_name").Scan(&listName).Error; err != nil {
		zap.L().Error("GetListDetailByID failed", zap.Error(err))
		return "", err
	}
	return
}
