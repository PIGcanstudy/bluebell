package controller

type MyCode int64

const (
	ResponseSuccess  MyCode = 1000
	InvalidParams    MyCode = 1001
	UserExist        MyCode = 1002
	PasswordNotMatch MyCode = 1003
	CodeServerBusy   MyCode = 1004
	PasswordError    MyCode = 1005
	UserNotExist     MyCode = 1006
	LoginFaild       MyCode = 1007
	InvalidToken     MyCode = 1008
	NotLogin         MyCode = 1009
	VoteRepated      MyCode = 1010
	DeletePostFaild  MyCode = 1011
)

var msgFlags = map[MyCode]string{
	ResponseSuccess:  "success",
	InvalidParams:    "请求参数错误",
	UserExist:        "用户名重复",
	PasswordNotMatch: "密码不匹配",
	CodeServerBusy:   "服务繁忙",
	PasswordError:    "密码错误",
	UserNotExist:     "用户不存在",
	LoginFaild:       "用户名或密码错误",
	InvalidToken:     "无效的token",
	NotLogin:         "未登录",
	VoteRepated:      "重复投相同的票",
	DeletePostFaild:  "删除帖子失败",
}

func (c MyCode) Msg() string {
	msg, ok := msgFlags[c]
	if ok {
		return msg
	}
	return msgFlags[CodeServerBusy]
}
