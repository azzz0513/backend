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
	err = DB.Table("users").Where("username=?", user.Username).First(&user).Error
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

// UpdateUser 更新用户数据
func UpdateUser(u *models.UpdateUser) (err error) {
	if err = DB.Table("users").Where("id=?", u.UserID).Omit("username", "email").Updates(u).Error; err != nil {
		zap.L().Error("update user failed", zap.Error(err))
		return
	}
	return
}

// ChangePassword 修改用户密码
func ChangePassword(u *models.ChangePassword) (err error) {
	// 使用事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检验用户旧密码是否正确
	oPassword := u.OldPassword
	var storedPassword string
	if err = tx.Table("users").Where("user_id=?", u.UserID).Select("password").Scan(&storedPassword).Error; err != nil {
		tx.Rollback()
		zap.L().Error("get user password failed", zap.Error(err))
		return
	}
	password := encryptPassword(oPassword)
	if password != storedPassword {
		tx.Rollback()
		return ErrorInvalidPassword
	}

	// 修改用户密码
	newPassword := encryptPassword(u.NewPassword)
	if err = tx.Table("users").Where("user_id=?", u.UserID).Select("password").Update("password", newPassword).Error; err != nil {
		tx.Rollback()
		zap.L().Error("update user password failed", zap.Error(err))
		return
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		zap.L().Error("事务提交失败",
			zap.Int64("user_id", u.UserID),
			zap.Error(err))
		return
	}
	return
}
