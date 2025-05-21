package models

// 定义请求的参数结构体

// ParamSignUp 定义注册请求的参数结构体
type ParamSignUp struct {
	Username   string `json:"username" binding:"required"` // 用户名，必填字段
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`                     // 用户密码，必填字段
	RePassword string `json:"re_password" binding:"required,eqfield=Password"` //  确认密码，必填字段
}

// ParamLogin 定义登录请求的参数结构体
type ParamLogin struct {
	Username string `json:"username" binding:"required"` // 用户名，必填字段
	Password string `json:"password" binding:"required"` // 用户密码，必填字段
}
