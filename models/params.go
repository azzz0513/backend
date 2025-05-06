package models

// 定义请求的参数结构体

const (
	OrderTime  = "time"
	OrderScore = "score"
)

// ParamSignUp 定义注册请求的参数结构体
type ParamSignUp struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password"`
}

// ParamLogin 定义登录请求的参数结构体
type ParamLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
