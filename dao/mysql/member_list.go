package mysql

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strings"
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
func AddMember(m *models.UpdateMember) (err error) {
	// 开启事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 插入成员记录
	if err = tx.Table("list_participants").
		Select("list_id", "user_id").
		Create(m).Error; err != nil {
		tx.Rollback()
		// 处理唯一约束冲突
		if strings.Contains(err.Error(), "Duplicate entry") {
			zap.L().Warn("重复添加成员",
				zap.Int64("list_id", m.ListID),
				zap.Int64("member_id", m.MemberID))
			return ErrorUserExist
		}
		zap.L().Error("插入成员记录失败", zap.Error(err))
		return
	}

	// 2. 更新计数
	result := tx.Table("member_list").
		Where("list_id = ?", m.ListID).
		Update("member_count", gorm.Expr("member_count + 1"))

	if err = result.Error; err != nil {
		tx.Rollback()
		zap.L().Error("更新成员计数失败",
			zap.Int64("list_id", m.ListID),
			zap.Error(err))
		return
	}

	// 验证影响行数
	if result.RowsAffected == 0 {
		tx.Rollback()
		zap.L().Warn("目标名单不存在",
			zap.Int64("list_id", m.ListID))
		return ErrorListNotFound
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败",
			zap.Int64("list_id", m.ListID),
			zap.Error(err))
		return ErrorCommitFailed
	}
	return
}

// DeleteMember 处理从成员列表删除指定成员的数据库交互
func DeleteMember(m *models.UpdateMember) (err error) {
	// 开启事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 删除成员记录
	result := tx.Table("list_participants").
		Where("list_id = ? AND user_id = ?", m.ListID, m.MemberID).
		Delete(nil)
	if err = result.Error; err != nil {
		tx.Rollback()
		zap.L().Error("删除成员记录失败",
			zap.Int64("list_id", m.ListID),
			zap.Int64("user_id", m.MemberID),
			zap.Error(err))
		zap.L().Error("删除成员记录失败", zap.Error(err))
		return
	}

	// 检查是否实际删除了记录
	if result.RowsAffected == 0 {
		tx.Rollback()
		zap.L().Warn("成员不存在",
			zap.Int64("list_id", m.ListID),
			zap.Int64("user_id", m.MemberID))
		return ErrorUserNotExist
	}

	// 2. 更新计数
	update := tx.Table("member_list").
		Where("list_id = ?", m.ListID).
		Update("member_count", gorm.Expr("member_count - 1"))

	if err = update.Error; err != nil {
		tx.Rollback()
		zap.L().Error("更新成员计数失败",
			zap.Int64("list_id", m.ListID),
			zap.Error(err))
		return
	}

	// 验证列表是否存在
	if result.RowsAffected == 0 {
		tx.Rollback()
		zap.L().Warn("目标名单不存在",
			zap.Int64("list_id", m.ListID))
		return ErrorListNotFound
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败",
			zap.Int64("list_id", m.ListID),
			zap.Error(err))
		return ErrorCommitFailed
	}
	return
}

// GetListList 处理获取当前啊用户创建的成员列表的数据库交互
func GetListList(authorID, page, size int64) (data []*models.MemberList, err error) {
	data = make([]*models.MemberList, 0, 2)
	if err = DB.Table("member_list").
		Where("author_id=?", authorID).
		Select("list_id", "author_id", "list_name", "create_time", "member_count").
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
