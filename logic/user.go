package logic

import (
	"bytes"
	"crypto/tls"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
	"html/template"
	"web_app/dao/mysql"
	"web_app/models"
	"web_app/pkg/jwt"
	"web_app/pkg/snowflake"
	"web_app/settings"
)

// SignUp 用户注册
func SignUp(p *models.ParamSignUp) (err error) {
	// 1.判断用户存不存在
	if err = mysql.CheckUserExist(p.Username); err != nil {
		// 数据库查询出错
		zap.L().Error("mysql.CheckUserExist", zap.Error(err))
		return
	}
	if err = mysql.CheckEmailExist(p.Email); err != nil {
		zap.L().Error("mysql.CheckEmailExist", zap.Error(err))
		return
	}
	// 2.生成UID
	userID := snowflake.GenID()
	// 3.构造一个User实例
	user := &models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
		Email:    p.Email,
	}
	// 4.保存进数据库
	return mysql.InsertUser(user)
}

// Login 用户登录
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

// GetUserDetail 获取用户详情
func GetUserDetail(userID int64) (userDetail *models.UserDetail, err error) {
	userDetail, err = mysql.GetUserDetail(userID)
	if err != nil {
		zap.L().Error("mysql.GetUserDetail failed", zap.Error(err))
		return nil, err
	}
	return
}

// UpdateUser 修改用户数据
func UpdateUser(u *models.UpdateUser) (err error) {
	if err = mysql.CheckUserExist(u.UserName); err != nil {
		// 数据库查询出错
		zap.L().Error("mysql.CheckUserExist failed", zap.Error(err))
		return
	}
	return mysql.UpdateUser(u)
}

// ChangePassword 修改用户密码
func ChangePassword(u *models.ChangePassword) (err error) {
	if err = mysql.ChangePassword(u); err != nil {
		return
	}
	return
}

// FindPassword 找回用户密码
func FindPassword(e *models.FindPassword) (err error) {
	// 获取指定邮箱对应的用户
	user, err := mysql.CheckEmail(e.Email)
	if err != nil {
		zap.L().Error("mysql.CheckEmail failed", zap.Error(err))
		return
	}
	// 生成重置令牌
	token, err := jwt.GenToken(user.UserID)
	if err != nil {
		return
	}
	// 发送邮件
	go sendResetEmail(e.Email, token)

	return
}

func sendResetEmail(email, token string) (err error) {
	// 从配置文件中获取邮件配置信息
	cfg := settings.Conf.EmailConfig

	// 渲染模板
	tpl, _ := template.ParseFiles("templates/reset_email.html")
	var body bytes.Buffer
	err = tpl.Execute(&body, struct{ Token string }{Token: token})
	if err != nil {
		zap.L().Error("template.Execute failed", zap.Error(err))
		return
	}

	// 构建邮件内容
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.User)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "密码重置请求")
	m.SetBody("text/html", body.String())

	// 创建Dialer
	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Password)

	// 设置 TLS 配置
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         cfg.Host,
	}

	// 发送邮件
	if err = d.DialAndSend(m); err != nil {
		zap.L().Error("发送邮件失败", zap.Error(err))
		return
	}

	return
}

// ResetPassword 重置用户密码
func ResetPassword(u *models.ResetPassword) (err error) {
	// 判断用户是否存在
	if err = mysql.CheckUserExist2(u.UserName); err != nil {
		zap.L().Error("mysql.CheckUserExist failed", zap.Error(err))
		return
	}
	// 重置密码
	if err = mysql.ResetPassword(u); err != nil {
		zap.L().Error("mysql.ResetPassword failed", zap.Error(err))
		return
	}
	return
}
