package mysql

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"math"
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
		"checkin_id", "author_id", "type_id", "way_id", "title", "content", "list_id",
	}
	// 根据list_name获取list_id
	if err = tx.Table("member_list").Where("list_name = ? AND author_id = ?", ck.ListName, ck.AuthorID).Select("list_id").Scan(&ck.ListID).Error; err != nil {
		tx.Rollback()
		zap.L().Error("获取列表id失败", zap.Error(err))
		return
	}
	switch ck.WayID {
	case 1:
		createFields = append(createFields, "password")
	case 2:
		break
	default:
		return errors.New("不支持的打卡方式")
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

// DeleteCheckin 根据checkin_id删除指定的打卡活动id
func DeleteCheckin(checkinID int64) (err error) {
	if err = DB.Table("checkins").Unscoped().Where("checkin_id = ?", checkinID).Delete(nil).Error; err != nil {
		zap.L().Error("删除打卡活动失败", zap.Error(err))
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

func GetDuration(checkinId int64) (duration uint, err error) {
	c := new(models.CheckinTime)
	if err = DB.Table("checkins").Where("checkin_id = ?", checkinId).Select("start_time", "duration_minutes").Scan(&c).Error; err != nil {
		zap.L().Error("获取活动时间失败", zap.Error(err))
		return 0, err
	}

	// 计算活动结束时间
	endTime := c.StartTime.Add(time.Duration(c.DurationMinutes) * time.Minute)

	// 获取当前时间
	currentTime := time.Now()

	// 计算剩余时间（分钟）
	oDuration := endTime.Sub(currentTime)
	duration = uint(math.Ceil(oDuration.Minutes()))

	// 如果活动已结束，返回0
	if oDuration < 0 {
		duration = 0
	}
	return
}

// GetCheckinMsg 根据活动id获取活动的基础信息
func GetCheckinMsg(id int64) (data *models.Checkin, err error) {
	data = new(models.Checkin)
	if err = DB.Table("checkins").
		Where("checkin_id = ?", id).
		Select("checkin_id", "author_id", "title", "content", "list_id", "type_id", "way_id", "status", "create_time", "update_time", "start_time", "duration_minutes", "end_date", "start_date", "daily_deadline").
		Scan(&data).Error; err != nil {
		zap.L().Error("get checkin fail", zap.Error(err))
		return
	}
	return
}

// UnCheckedMember 根据活动id获取未完成打卡的用户列表
func UnCheckedMember(id, page, size int64) (count int, members []*models.UserEasyDetail, err error) {
	// 使用事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询未完成的成员列表
	members = make([]*models.UserEasyDetail, 0)
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
		tx.Rollback()
		zap.L().Error("获取未完成成员列表失败", zap.Error(err))
		return
	}
	// 查询未完成人数
	queryCompleted := `
        SELECT COUNT(DISTINCT cr.user_id)
        FROM checkins c
        JOIN checkin_records cr ON c.checkin_id = cr.checkin_id
        WHERE 
            c.checkin_id = ?
            AND cr.is_checked = 0`
	if err = tx.Raw(queryCompleted, id).Scan(&count).Error; err != nil {
		tx.Rollback()
		zap.L().Error("获取未完成成员人数失败", zap.Error(err))
		return
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败",
			zap.Int64("checkin_id", id),
			zap.Error(err))
		return
	}

	return
}

// CheckedMember 根据活动id获取已完成打卡的用户列表
func CheckedMember(id, page, size int64) (count int, members []*models.UserEasyDetail, err error) {
	// 使用事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询已完成的成员列表
	members = make([]*models.UserEasyDetail, 0)
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
            AND (cr.is_checked = 1 OR cr.is_checked IS NULL)`
	if err = tx.Raw(queryUncompleted, id).Offset(int((page - 1) * size)).Limit(int(size)).Scan(&members).Error; err != nil {
		tx.Rollback()
		zap.L().Error("获取已完成成员列表失败", zap.Error(err))
		return
	}
	// 查询未完成人数
	queryCompleted := `
        SELECT COUNT(DISTINCT cr.user_id)
        FROM checkins c
        JOIN checkin_records cr ON c.checkin_id = cr.checkin_id
        WHERE 
            c.checkin_id = ?
            AND cr.is_checked = 1`
	if err = tx.Raw(queryCompleted, id).Scan(&count).Error; err != nil {
		tx.Rollback()
		zap.L().Error("获取已完成成员人数失败", zap.Error(err))
		return
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败",
			zap.Int64("checkin_id", id),
			zap.Error(err))
		return
	}

	return
}

// CheckCheckinPassword 检查用户传进的验证码与活动指定验证码是否相符
func CheckCheckinPassword(checkinID int64, password string) (ok bool, err error) {
	var realPassword string
	if err = DB.Table("checkins").Where("checkin_id = ? AND status = 1", checkinID).Select("password").Scan(&realPassword).Error; err != nil {
		zap.L().Error("检查打卡活动验证码失败", zap.Error(err))
		return false, err
	}
	if password != realPassword {
		return false, nil
	}
	return true, nil
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
                AND cr.user_id = ?
				AND cr.is_checked = 1`, userID).
		Joins(`
            INNER JOIN list_participants lp 
                ON c.list_id = lp.list_id 
                AND lp.user_id = ?`, userID).
		Where("cr.is_checked = 1 OR c.status = 0").
		Order("cr.check_time DESC").
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Scan(&data).Error; err != nil {
		zap.L().Error("获取当前用户的已打卡的打卡活动历史记录失败", zap.Error(err))
		return nil, err
	}
	return
}

// GetStatistics 获取当前活动的统计数据
func GetStatistics(checkinID int64, statsType string) (data []*models.MsgStatistics, err error) {
	// 开启事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Table("checkin_stats").
		Where("checkin_id = ? AND period_type = ?", checkinID, statsType).
		Select("checkin_id", "user_id", "period_type", "checkin_count", "last_checkin_time").
		Scan(&data).Error; err != nil {
		tx.Rollback()
		zap.L().Error("获取统计数据失败", zap.Error(err))
		return nil, err
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败",
			zap.Int64("checkin_id", checkinID),
			zap.String("statsType", statsType),
			zap.Error(err))
		return
	}
	return
}

// GetUserInfo 检验用户活动权限并返回用户信息
func GetUserInfo(checkinID, userID int64) (data *models.User, err error) {
	data = new(models.User)
	var exists bool
	if err = DB.Table("checkin_records").Where("checkin_id = ? AND user_id = ?", checkinID, userID).Select("count(*) > 0").Scan(&exists).Error; err != nil {
		zap.L().Error("检查用户活动权限失败", zap.Error(err))
		return nil, err
	}
	if !exists {
		zap.L().Error("用户无活动权限", zap.Error(err))
		return nil, errors.New("用户无活动权限")
	}
	if err = DB.Table("users").
		Where("id = ?", userID).
		Select("user_id", "username").
		Scan(&data).Error; err != nil {
		zap.L().Error("获取用户信息失败", zap.Error(err))
		return nil, err
	}
	return
}

// GetRangeByID 获取定位范围
func GetRangeByID(checkinID int64) (lat, lng, radius float64, err error) {
	p := new(models.MsgPosition)
	if err = DB.Table("checkins").
		Where("checkin_id = ?", checkinID).
		Select("latitude", "longitude", "radius").Scan(&p).Error; err != nil {
		zap.L().Error("获取打卡活动范围失败", zap.Error(err))
		return 0, 0, 0, err
	}
	lat, lng, radius = p.Lat, p.Lng, p.Radius
	return
}
