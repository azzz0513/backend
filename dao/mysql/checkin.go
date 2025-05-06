package mysql

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
	"web_app/models"
)

// CreateCheckin 创建打卡活动
func CreateCheckin(ck *models.Checkin) (err error) {
	// 开启事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 创建活动主体
	createFields := []string{
		"checkin_id", "author_id", "type_id", "way_id",
		"status", "title", "content", "list_id", "password",
	}

	// 根据type_id判断打卡活动类型
	// 根据类型添加专属字段
	switch ck.TypeID {
	case 1: // 一次性签到
		createFields = append(createFields, "start_time", "duration_minutes")
	case 2: // 长期考勤
		createFields = append(createFields, "start_date", "end_date", "daily_deadline")
	default:
		return errors.New("不支持的活动类型")
	}

	if err = tx.Table("checkins").Select(createFields).Create(ck).Error; err != nil {
		tx.Rollback()
		zap.L().Error("创建活动失败",
			zap.Error(err),
			zap.Any("checkin", ck))
		return
	}

	// 2. 初始化打卡记录
	// 获取名单中的参与者
	var participantIDs []struct {
		UserID int64 `gorm:"column:user_id"`
	}
	if err := tx.Table("list_participants").
		Select("user_id").
		Where("list_id = ?", ck.ListID).
		Find(&participantIDs).Error; err != nil {
		tx.Rollback()
		zap.L().Error("获取参与名单失败",
			zap.Int64("list_id", ck.ListID),
			zap.Error(err))
		return fmt.Errorf("获取参与名单失败")
	}

	// 构造初始化记录
	records := make([]models.CheckinRecord, 0, len(participantIDs))
	for _, p := range participantIDs {
		records = append(records, models.CheckinRecord{
			CheckinID: ck.ID,
			UserID:    p.UserID,
			ListID:    ck.ListID,
			IsChecked: 0,
			CheckTime: nil,
		})
	}

	// 批量插入（每批500条）
	if len(records) > 0 {
		if err = tx.CreateInBatches(records, 500).Error; err != nil {
			tx.Rollback()
			zap.L().Error("初始化打卡记录失败",
				zap.Int("records", len(records)),
				zap.Error(err))
			return
		}
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		zap.L().Error("创建打卡活动事务提交失败",
			zap.Error(err))
		return
	}
	return
}

// GetTypeDetailByID 根据type_id获取对应的打卡活动类型信息
func GetTypeDetailByID(typeID int64) (detail string, err error) {
	if err = DB.Table("type").Where("type_id = ?", typeID).Select("type_name").Scan(detail).Error; err != nil {
		zap.L().Warn("查询的打卡活动类型不存在", zap.Error(err))
		return "", err
	}
	return
}

// GetWayDetailByID 根据way_id获取对应的打卡方式信息
func GetWayDetailByID(wayID int64) (detail string, err error) {
	if err = DB.Table("way").Where("way_id = ?", wayID).Select("way_name").Scan(detail).Error; err != nil {
		zap.L().Warn("查询的打卡方式不存在", zap.Error(err))
		return "", err
	}
	return
}

// GetCheckinMsg 根据活动id获取活动的基础信息
func GetCheckinMsg(id int64) (data *models.Checkin, err error) {
	data = new(models.Checkin)
	if err = DB.Table("checkins").
		Where("checkin_id = ?", id).Select("author_id", "title", "content", "list_id", "type_id", "way_id", "status").
		Scan(&data).Error; err != nil {
		zap.L().Error("get checkin fail", zap.Error(err))
		return
	}
	if data.Status != 0 {
		if data.TypeID == 1 {
			// 当前打卡活动为一次性打卡活动
			if err = DB.Table("checkins").Where("checkin_id = ?", id).Select("start_time", "duration_minutes").Scan(&data).Error; err != nil {
				zap.L().Error("get checkin fail", zap.Error(err))
				return
			}
		} else if data.TypeID == 2 {
			// 当前打卡活动为长期考勤活动
			if err = DB.Table("checkins").Where("checkin_id = ?", id).Select("end_time", "start_time", "daily_deadline").Scan(&data).Error; err != nil {
				zap.L().Error("get checkin fail", zap.Error(err))
				return
			}
		}
	}
	return
}

// CheckMember 根据活动id获取未完成打卡的用户列表
func CheckMember(id int64) (members []*models.UserDetail, err error) {
	members = make([]*models.UserDetail, 0)
	query := `
        SELECT 
            u.user_id,
            u.username
        FROM checkins c
        JOIN list_participants lp ON c.list_id = lp.list_id
        JOIN users u ON lp.user_id = u.user_id
        LEFT JOIN checkin_records cr 
            ON cr.checkin_id = c.checkin_id 
            AND cr.user_id = u.user_id
        WHERE 
            c.checkin_id = ? 
            AND (cr.is_checked = 0 OR cr.is_checked IS NULL)`
	err = DB.Raw(query, id).Scan(&members).Error
	if err != nil {
		zap.L().Error("get checkin fail", zap.Error(err))
		return
	}
	return
}

// Participate 更新用户打卡状态
func Participate(userID, checkinID int64) (err error) {
	// 开启事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1.原子性获取活动状态
	// 构建获取活动状态的结构体
	var checkin struct {
		Status    int32
		ListID    int64
		StartTime time.Time
	}
	if err = tx.Table("checkins").
		Select("status, list_id, start_time").
		Where("checkin_id = ?", checkinID).
		First(&checkin).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			zap.L().Warn("活动不存在",
				zap.Int64("checkin_id", checkinID),
				zap.Int64("user_id", userID))
			return ErrorActivityNotFound
		}
		zap.L().Error("获取活动状态失败",
			zap.Int64("checkin_id", checkinID),
			zap.Error(err))
		return
	}

	// 2.判断活动是否有效
	if checkin.Status != 1 {
		tx.Rollback()
		zap.L().Info("活动已过期或未开始",
			zap.Int64("checkin_id", checkinID),
			zap.Int64("user_id", userID))
		return ErrorActivityExpire
	}

	// 3.更新打卡状态
	result := tx.Table("checkin_records").
		Where("checkin_id = ? AND user_id = ? AND is_checked = 0", checkinID, userID).
		Updates(map[string]interface{}{
			"is_checked": 1,
			"check_time": time.Now(),
		})

	if result.Error != nil {
		tx.Rollback()
		zap.L().Error("更新打卡记录失败",
			zap.Int64("checkin_id", checkinID),
			zap.Int64("user_id", userID),
			zap.Error(result.Error))
		return result.Error
	}

	// 4. 检查是否成功更新
	if result.RowsAffected == 0 {
		tx.Rollback()
		zap.L().Info("用户重复打卡",
			zap.Int64("checkin_id", checkinID),
			zap.Int64("user_id", userID))
		return ErrorChecked
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败",
			zap.Int64("checkin_id", checkinID),
			zap.Int64("user_id", userID),
			zap.Error(err))
		return
	}
	return
}

// GetCheckinList 获取当前用户需要参与的打卡活动的列表
func GetCheckinList(userID, page, size int64) (data []*models.Checkin, err error) {
	data = make([]*models.Checkin, 0)
	if err = DB.Table("checkins c").
		Select(`c.checkin_id,c.author_id,c.title,c.content,c.list_id,c.type_id,c.way_id,c.status,cr.is_checked AS user_checked`).
		Joins(`
            INNER JOIN list_participants lp 
                ON c.list_id = lp.list_id
            LEFT JOIN checkin_records cr 
                ON cr.checkin_id = c.checkin_id 
                AND cr.user_id = ?`, userID).
		Where("lp.user_id = ? AND c.status = 1", userID).
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Scan(&data).Error; err != nil {
		zap.L().Error("participant get checkin fail", zap.Error(err))
		return nil, err
	}
	return
}

// GetCreatedCheckinList 获取当前用户创建的打卡活动的列表
func GetCreatedCheckinList(userID, page, size int64) (data []*models.Checkin, err error) {
	data = make([]*models.Checkin, 0)
	if err = DB.Table("checkins").
		Where("author_id = ?", userID).
		Select("checkin_id", "author_id", "title", "content", "list_id", "type_id", "way_id", "status").
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Scan(&data).Error; err != nil {
		zap.L().Error("creator get checkin fail", zap.Error(err))
		return nil, err
	}
	return
}
