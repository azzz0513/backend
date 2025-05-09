package mysql

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

// CheckDisposable 检查一次性打卡活动的状态
func CheckDisposable() (err error) {
	// 使用事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查状态
	result := tx.Table("checkins").
		Where("type_id = 1").
		Update("status", gorm.Expr(`
                CASE 
                    WHEN NOW() BETWEEN start_time 
                        AND start_time + INTERVAL duration_minutes MINUTE 
                    THEN 1 
                    ELSE 0 
                END`))

	if err = result.Error; err != nil {
		tx.Rollback()
		zap.L().Error("检查更新一次性打卡活动状态失败", zap.Error(err))
		return
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败", zap.Error(err))
		return ErrorCommitFailed
	}

	zap.L().Info("状态更新完成",
		zap.Int64("影响行数", result.RowsAffected))
	return
}

// CheckLongTerm 检查长期考勤活动的状态
func CheckLongTerm() (err error) {
	// 使用事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查状态
	result := tx.Table("checkins").
		Where("type_id = 2").
		Update("status", gorm.Expr(`
                CASE 
                    WHEN NOW() BETWEEN start_date AND end_date
                        AND TIME(NOW()) < daily_deadline
                    THEN 1 
                    ELSE 0 
                END`))

	if err = result.Error; err != nil {
		tx.Rollback()
		zap.L().Error("检查更新长期考勤活动状态失败", zap.Error(err))
		return
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败", zap.Error(err))
		return ErrorCommitFailed
	}

	zap.L().Info("状态更新完成",
		zap.Int64("影响行数", result.RowsAffected))
	return
}

// FullCheckDisposable 全量检查一次性活动状态
func FullCheckDisposable() (err error) {
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 使用数据库当前时间计算
	result := tx.Table("checkins").
		Where("type_id = 1").
		Update("status", gorm.Expr(`
                CASE 
                    WHEN NOW() BETWEEN start_time 
                        AND start_time + INTERVAL duration_minutes MINUTE 
                    THEN 1 
                    ELSE 0 
                END`))

	if err = result.Error; err != nil {
		tx.Rollback()
		zap.L().Error("全量一次性活动检查失败", zap.Error(err))
		return
	}

	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败",
			zap.Error(err),
			zap.Int64("affected", result.RowsAffected))
		return ErrorCommitFailed
	}

	zap.L().Info("全量一次性活动检查完成",
		zap.Int64("影响行数", result.RowsAffected))
	return
}

// FullCheckLongTerm 全量检查长期活动状态
func FullCheckLongTerm() (err error) {
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取当前日期
	var currentDate string
	if err := tx.Raw("SELECT DATE(NOW())").Scan(&currentDate).Error; err != nil {
		return err
	}

	// 批量更新逻辑并获得被更新的活动ID
	var activeCheckIDs []int64
	result := tx.Table("checkins").
		Where("type_id = 2").
		Update("status", gorm.Expr(`
                CASE 
                    WHEN NOW() BETWEEN start_date AND end_date
                        AND TIME(NOW()) < daily_deadline
                    THEN 1 
                    ELSE 0 
                END`)).
		Where("status != ?", gorm.Expr(`
                CASE 
                    WHEN NOW() BETWEEN start_date AND end_date
                        AND TIME(NOW()) < daily_deadline
                    THEN 1 
                    ELSE 0 
                END`)).Scan(&activeCheckIDs)

	if err = result.Error; err != nil {
		zap.L().Error("全量长期活动检查失败",
			zap.Error(err),
			zap.String("current_date", currentDate))
		return
	}

	// 重置打卡记录（仅处理有效的长期活动）
	if len(activeCheckIDs) > 0 {
		resetResult := tx.Table("checkins").
			Table("checkin_records").
			Where("checkin_id IN (?) AND DATE(check_time) = ? AND is_checked = 1", activeCheckIDs, currentDate).
			Set("is_checked = ?,check_time = NULL", 1)

		if err = resetResult.Error; err != nil {
			tx.Rollback()
			zap.L().Error("打卡记录重置失败",
				zap.Error(err),
				zap.Int("影响活动数", len(activeCheckIDs)))
			return
		}
	}

	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败",
			zap.Error(err),
			zap.Int64("affected", result.RowsAffected))
		return ErrorCommitFailed
	}

	zap.L().Info("全量长期活动检查完成",
		zap.Int64("影响行数", result.RowsAffected),
		zap.Int("影响活动数", len(activeCheckIDs)),
		zap.String("当前日期", currentDate))
	return
}

// CleanExpiredCheckins 清理过期活动（保留30天）
func CleanExpiredCheckins() (err error) {
	// 保留最近30天已结束的活动
	result := DB.Exec(`
        DELETE FROM checkins 
        WHERE status = 0 
          AND (type_id = 1 AND NOW() > start_time + INTERVAL duration_minutes MINUTE + INTERVAL 30 DAY)
          OR (type_id = 2 AND CURDATE() > end_date + INTERVAL 30 DAY)
    `)

	if err = result.Error; err != nil {
		zap.L().Error("清理过期活动失败",
			zap.Error(err),
			zap.String("SQL", "CleanExpiredCheckins"))
		return
	}

	zap.L().Info("过期活动清理完成",
		zap.Int64("删除记录数", result.RowsAffected))
	return
}

// DailyStat 每日进行数据统计
func DailyStat() (err error) {
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 统计昨日数据
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 按天统计
	if err = tx.Exec(`
        INSERT INTO checkin_stats 
        (checkin_id, user_id, period_type, period_start, period_end, checkin_count, last_checkin_time)
        SELECT 
            r.checkin_id,
            r.user_id,
            'day' AS period_type,
            DATE(r.check_time) AS period_start,
            DATE(r.check_time) AS period_end,
            COUNT(CASE WHEN r.is_checked = 1 THEN 1 ELSE 0 END) AS checkin_count,
            MAX(CASE WHEN r.is_checked = 1 THEN r.check_time ELSE NULL END) AS last_checkin_time
        FROM checkin_records r
        WHERE EXISTS(
            SELECT 1 FROM checkins c
        	WHERE c.checkin_id = r.checkin_id
        	AND c.type_id = 2
        	AND ? BETWEEN c.start_date AND c.end_date
        )
        GROUP BY r.checkin_id, r.user_id, DATE(r.check_time)
        ON DUPLICATE KEY UPDATE 
            checkin_count = VALUES(checkin_count),
            last_checkin_time = VALUES(last_checkin_time)
    `, yesterday, yesterday).Error; err != nil {
		tx.Rollback()
		zap.L().Error("按日统计数据出错", zap.Error(err))
		return
	}

	// 按周统计（每周日触发）
	if time.Now().Weekday() == time.Sunday {
		if err = tx.Exec(`
            INSERT INTO checkin_stats 
            (checkin_id, user_id, period_type, period_start, period_end, checkin_count, last_checkin_time)
            SELECT 
                r.checkin_id,
                r.user_id,
                'week' AS period_type,
                DATE_SUB(DATE(r.check_time), INTERVAL WEEKDAY(r.check_time) DAY) AS period_start,
                DATE_ADD(DATE(r.check_time), INTERVAL (6 - WEEKDAY(r.check_time)) DAY) AS period_end,
                SUM(CASE WHEN r.is_checked = 1 THEN 1 ELSE 0 END) AS checkin_count,
                MAX(CASE WHEN r.is_checked = 1 THEN r.check_time ELSE NULL END) AS last_checkin_time
            FROM checkin_records r
            WHERE EXISTS(
                SELECT 1 FROM checkins c
                WHERE c.checkin_id = r.checkin_id
                	AND c.type_id = 2
                	AND r.check_time BETWEEN c.start_date AND c.end_date
            )
            AND YEARWEEK(r.check_time) = YEARWEEK(NOW() - INTERVAL 1 WEEK)
            GROUP BY r.checkin_id, r.user_id, YEARWEEK(r.check_time)
        `).Error; err != nil {
			tx.Rollback()
			zap.L().Error("按周统计数据出错", zap.Error(err))
			return
		}
	}

	// 每月最后一天统计
	if time.Now().AddDate(0, 1, -time.Now().Day()).Day() == time.Now().Day() {
		if err = tx.Exec(`
            INSERT INTO checkin_stats 
            (checkin_id, user_id, period_type, period_start, period_end, checkin_count, last_checkin_time)
            SELECT 
                r.checkin_id,
                r.user_id,
                'month' AS period_type,
                DATE_FORMAT(r.check_time, '%Y-%m-01') AS period_start,
                LAST_DAY(r.check_time) AS period_end,
                SUM(CASE WHEN r.is_checked = 1 THEN 1 ELSE 0 END) AS checkin_count,
                MAX(CASE WHEN r.is_checked = 1 THEN r.check_time ELSE NULL END) AS last_checkin_time
            FROM checkin_records r
            WHERE EXISTS (
    			SELECT 1 FROM checkins c
    			WHERE c.id = r.checkin_id
      				AND c.type_id = 2
      				AND r.check_time BETWEEN c.start_date AND c.end_date
			)
  			AND DATE_FORMAT(r.check_time, '%Y-%m') = DATE_FORMAT(NOW() - INTERVAL 1 MONTH, '%Y-%m')
            GROUP BY r.checkin_id, r.user_id, DATE_FORMAT(r.check_time, '%Y-%m')
        `).Error; err != nil {
			tx.Rollback()
			zap.L().Error("按月统计数据失败", zap.Error(err))
			return
		}
	}

	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败",
			zap.Error(err))
		return ErrorCommitFailed
	}
	return
}
