package controller

type ResCode int64

const (
	CodeSuccess ResCode = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeServerBusy
	CodeNeedLogin
	CodeInvalidToken
	CodeNeedPassword
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess:         "业务处理成功",
	CodeInvalidParam:    "请求参数错误",
	CodeUserExist:       "用户名已存在",
	CodeUserNotExist:    "用户名不存在",
	CodeInvalidPassword: "用户名或密码错误",
	CodeServerBusy:      "服务繁忙",
	CodeNeedLogin:       "需要登录",
	CodeInvalidToken:    "无效的Token",
	CodeNeedPassword:    "需要正确的验证码",
}

func (code ResCode) Msg() string {
	msg, ok := codeMsgMap[code]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
