package statuschecker

import (
	"go.uber.org/zap"
	"time"
	"web_app/dao/mysql"
)

// CheckStatus 检查所有打卡活动的当前状态
func CheckStatus() {
	// 开机立即执行第一次检查
	check()
	zap.L().Info("开机状态检查", zap.Time("start_time", time.Now()))

	// 启动协程处理不同频率的检查
	go runMinuteCheck() // 每分钟检查
	go runDailyCheck()  // 每日凌晨检查
	go runDailyStat()   // 每日凌晨统计数据

	// 保持主协程运行
	select {}
}

func runMinuteCheck() {
	// 启动定时任务（每分钟执行）
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		check()
	}
}

func runDailyCheck() {
	for {
		// 计算到下一个凌晨的剩余时间
		now := time.Now()
		next := now.Add(24 * time.Hour)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 5, 0, 0, next.Location())
		duration := next.Sub(now)

		// 等待到目标时间
		<-time.After(duration)

		// 执行全量检查
		fullCheck()
	}
}

func check() {
	// 一次性签到活动状态更新
	if err := mysql.CheckDisposable(); err != nil {
		zap.L().Error("CheckDisposable failed", zap.Error(err))
		return
	}
	// 长期考勤活动状态更新
	if err := mysql.CheckLongTerm(); err != nil {
		zap.L().Error("CheckLongTerm failed", zap.Error(err))
		return
	}
}

func fullCheck() {
	zap.L().Info("开始全量状态检查")
	defer zap.L().Info("完成全量状态检查")
	// 全量更新一次性活动
	if err := mysql.FullCheckDisposable(); err != nil {
		zap.L().Error("FullCheckDisposable failed", zap.String("type", "disposable"), zap.Error(err))
	}
	// 全量更新长期活动
	if err := mysql.FullCheckLongTerm(); err != nil {
		zap.L().Error("FullCheckLongTerm failed", zap.String("type", "long_term"), zap.Error(err))
	}
	// 清理过期数据（可选）
	if err := mysql.CleanExpiredCheckins(); err != nil {
		zap.L().Error("CleanExpiredCheckins failed", zap.Error(err))
	}
}

func runDailyStat() {
	for {
		// 计算到下一个凌晨的剩余时间
		now := time.Now()
		next := now.Add(24 * time.Hour)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 5, 0, 0, next.Location())
		duration := next.Sub(now)

		// 等待到目标时间
		<-time.After(duration)

		// 执行数据统计
		if err := mysql.DailyStat(); err != nil {
			zap.L().Error("daily stat failed", zap.Error(err))
			return
		}
	}
}
