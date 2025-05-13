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
		"checkin_id", "author_id", "type_id", "way_id", "title", "content", "list_id", "password",
	}
	// 根据list_name获取list_id
	if err = tx.Table("member_list").Where("list_name = ? AND author_id = ?", ck.ListName, ck.AuthorID).Select("list_id").Scan(&ck.ListID).Error; err != nil {
		tx.Rollback()
		zap.L().Error("获取列表id失败", zap.Error(err))
		return
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
	if err = tx.Table("list_participants").
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
			CheckTime: time.Now(),
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
	if err = DB.Table("type").Where("type_id = ?", typeID).Select("type_name").Scan(&detail).Error; err != nil {
		zap.L().Warn("查询的打卡活动类型不存在", zap.Error(err))
		return "", err
	}
	return
}

// GetWayDetailByID 根据way_id获取对应的打卡方式信息
func GetWayDetailByID(wayID int64) (detail string, err error) {
	if err = DB.Table("way").Where("way_id = ?", wayID).Select("way_name").Scan(&detail).Error; err != nil {
		zap.L().Warn("查询的打卡方式不存在", zap.Error(err))
		return "", err
	}
	return
}

// GetCheckTime 获取打卡时间
func GetCheckTime(checkID, userID int64) (checkTime time.Time, err error) {
	if err = DB.Table("checkin_records").Where("checkin_id = ? AND user_id = ?", checkID, userID).Select("check_time").Scan(&checkTime).Error; err != nil {
		zap.L().Error("查询当前用户打卡时间失败", zap.Error(err))
		return time.Time{}, err
	}
	return
}

// GetCheckinMsg 根据活动id获取活动的基础信息
func GetCheckinMsg(id int64) (data *models.Checkin, err error) {
	data = new(models.Checkin)
	if err = DB.Table("checkins").
		Where("checkin_id = ?", id).Select("checkin_id", "author_id", "title", "content", "list_id", "type_id", "way_id", "status", "create_time", "update_time", "start_time", "duration_minutes", "end_date", "start_date", "daily_deadline").
		Scan(&data).Error; err != nil {
		zap.L().Error("get checkin fail", zap.Error(err))
		return
	}
	return
}

// CheckMember 根据活动id获取未完成打卡的用户列表
func CheckMember(id, page, size int64) (count int, members []*models.UserDetail, err error) {
	// 使用事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询未完成的成员列表
	members = make([]*models.UserDetail, 0)
	queryUncompleted := `
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
	if err = tx.Raw(queryUncompleted, id).Offset(int((page - 1) * size)).Limit(int(size)).Scan(&members).Error; err != nil {
		zap.L().Error("获取未完成成员列表失败", zap.Error(err))
		return
	}
	// 查询已完成人数
	queryCompleted := `
        SELECT COUNT(DISTINCT cr.user_id)
        FROM checkins c
        JOIN checkin_records cr ON c.checkin_id = cr.checkin_id
        WHERE 
            c.checkin_id = ?
            AND cr.is_checked = 1`
	if err = tx.Raw(queryCompleted, id).Scan(&count).Error; err != nil {
		zap.L().Error("获取已完成成员人数失败", zap.Error(err))
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

	// 1.获取活动状态
	// 构建获取活动状态的结构体
	checkin := new(models.Checkin)
	if err = tx.Table("checkins").
		Select("status, type_id, list_id, start_time").
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
		Select(`c.checkin_id,c.author_id,c.title,c.content,c.list_id,c.type_id,c.way_id,c.status,c.create_time,c.update_time,COALESCE(cr.is_checked, 0) AS user_checked`).
		Joins(`
            INNER JOIN list_participants lp 
                ON c.list_id = lp.list_id
            LEFT JOIN checkin_records cr 
                ON cr.checkin_id = c.checkin_id 
                AND cr.user_id = ?`, userID).
		Where(`
            lp.user_id = ? 
            AND c.status = 1
            AND (cr.is_checked IS NULL OR cr.is_checked = 0)
            AND (
                (c.type_id = 1 AND NOW() BETWEEN c.start_time AND c.start_time + INTERVAL c.duration_minutes MINUTE)
                OR
                (c.type_id = 2 AND NOW() BETWEEN c.start_date AND c.end_date
                        AND TIME(NOW()) < c.daily_deadline)
            )`, userID).
		Order("c.create_time DESC").
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Scan(&data).Error; err != nil {
		zap.L().Error("获取当前用户需要参与打卡的打卡活动列表失败", zap.Error(err))
		return nil, err
	}
	return
}

// GetCreatedCheckinList 获取当前用户创建的打卡活动的列表
func GetCreatedCheckinList(userID, page, size int64) (data []*models.Checkin, err error) {
	data = make([]*models.Checkin, 0)
	if err = DB.Table("checkins").
		Where("author_id = ?", userID).
		Select("checkin_id", "author_id", "title", "content", "list_id", "type_id", "way_id", "status", "create_time", "update_time").
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Scan(&data).Error; err != nil {
		zap.L().Error("creator get checkin fail", zap.Error(err))
		return nil, err
	}
	return
}

// GetHistoryList 获取当前用户参与过的打卡活动的历史记录列表
func GetHistoryList(userID, page, size int64) (data []*models.Checkin, err error) {
	data = make([]*models.Checkin, 0)
	if err = DB.Table("checkins c").
		Select(`c.checkin_id,c.author_id,c.title,c.content,c.list_id,c.type_id,c.way_id,c.create_time,c.update_time,c.status,1 AS user_checked`).
		Joins(`
            INNER JOIN checkin_records cr 
                ON cr.checkin_id = c.checkin_id 
                AND cr.user_id = ?`, userID).
		Joins(`
            INNER JOIN list_participants lp 
                ON c.list_id = lp.list_id 
                AND lp.user_id = ?`, userID).
		Where("c.status = 0").
		Order("cr.check_time DESC").
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Scan(&data).Error; err != nil {
		zap.L().Error("获取当前用户的已打卡的打卡活动历史记录失败", zap.Error(err))
		return nil, err
	}
	return
}
