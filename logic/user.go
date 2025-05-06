package logic

import (
	"web_app/dao/mysql"
	"web_app/models"
	"web_app/pkg/jwt"
	"web_app/pkg/snowflake"
)

func SignUp(p *models.ParamSignUp) (err error) {
	// 1.判断用户存不存在
	if err = mysql.CheckUserExist(p.Username); err != nil {
		// 数据库查询出错
		return
	}
	// 2.生成UID
	userID := snowflake.GenID()
	// 3.构造一个User实例
	user := &models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
	}
	// 4.保存进数据库
	return mysql.InsertUser(user)
}

func Login(p *models.ParamLogin) (user *models.User, err error) {
	user = &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	// 登录
	if err = mysql.Login(user); err != nil {
		return nil, err
	}
	// 生成JWT
	token, err := jwt.GenToken(user.UserID)
	if err != nil {
		return
	}
	user.Token = token
	return
}
