package mysql

import "errors"

var (
	ErrorUserExist        = errors.New("用户已存在")
	ErrorUserNotExist     = errors.New("用户不存在")
	ErrorInvalidPassword  = errors.New("用户名或密码错误")
	ErrorInvalidID        = errors.New("无效的ID")
	ErrorListNotFound     = errors.New("指定名单不存在")
	ErrorActivityNotFound = errors.New("活动不存在")
	ErrorActivityExpire   = errors.New("活动已过期")
	ErrorChecked          = errors.New("已打卡")
	ErrorCommitFailed     = errors.New("事务提交失败")
)
