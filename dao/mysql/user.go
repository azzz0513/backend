package mysql

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"go.uber.org/zap"
	"web_app/models"
)

/*
	把每一步数据库操作封装为函数
	等待logic层根据业务需求调用
*/

const secret = "shit.com"

// CheckUserExist 检查指定用户名的用户是否存在
func CheckUserExist(username string) (err error) {
	var count int
	if err := DB.Raw("select count(user_id) from users where username = ?", username).Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrorUserExist
	}
	return
}

// InsertUser 向数据库中插入一条新的用户记录
func InsertUser(user *models.User) (err error) {
	// 对密码进行加密
	oPassword := user.Password
	user.Password = encryptPassword(oPassword)
	// 执行SQL语句入库
	sqlStr := `insert into users (user_id,username,password) values (?,?,?)`
	err = DB.Exec(sqlStr, user.UserID, user.Username, user.Password).Error
	return
}

// 对密码进行加密
func encryptPassword(password string) string {
	h := md5.New()
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum([]byte(password)))
}

// Login 判断用户登录时输入的用户名是否存在以及密码是否正确
func Login(user *models.User) (err error) {
	oPassword := user.Password // 用户登录时输入的密码
	// 查询用户是否存在
	err = DB.Where("username=?", user.Username).First(&user).Error
	if errors.Is(err, sql.ErrNoRows) {
		return ErrorUserNotExist
	}
	if err != nil {
		// 查询数据库失败
		zap.L().Error("查询数据库失败", zap.Error(err))
		return
	}
	// 判断密码是否正确
	password := encryptPassword(oPassword)
	if password != user.Password {
		return ErrorInvalidPassword
	}
	return
}

// GetUserByID 根据id获取用户信息
func GetUserByID(uid int64) (user *models.User, err error) {
	user = new(models.User)
	if err = DB.Table("users").Where("user_id=?", uid).Select("user_id", "username").Scan(&user).Error; err != nil {
		err = ErrorInvalidID
	}
	return
}
