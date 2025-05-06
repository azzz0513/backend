package mysql

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"web_app/settings"
)

var DB *gorm.DB

// Init 初始化MySQL数据库连接
func Init(cfg *settings.MySQLConfig) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
	)
	// 连接MySQL，获得DB类型实例，用于后面的数据库读写操作
	db, err := gorm.Open(mysql.Open(dsn))
	// 连接失败
	if err != nil {
		zap.L().Error("数据库连接失败", zap.Error(err))
		return
	}
	// 连接成功
	DB = db
	return
}
