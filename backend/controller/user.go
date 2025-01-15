package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

func SignupHandler(c *gin.Context) {
	// 获取前端传来的参数
	var data models.RegisterForm

	// 验证参数是否符合注册表格的要求
	if err := c.ShouldBindJSON(&data); err != nil {
		zap.L().Error("注册时参数错误", zap.Error(err))
		code := InvalidParams
		ResponseError(c, code)
		return
	}

	// 验证确认密码和密码是否一致
	if data.Password != data.Confirm {
		code := PasswordNotMatch
		ResponseError(c, code)
		return
	}

	//  开始处理业务注册逻辑
	if isSuccess := logic.Signup(data); !isSuccess {
		code := UserExist
		ResponseError(c, code)
		return
	}

	// 已经注册完毕了
	ResponseSuccessed(c, "注册成功")
}

func SigninHandler(c *gin.Context) {
	// 获取前端传来的参数
	var data models.LoginForm

	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("登录时参数错误", err)
		return
	}

	// 返回-1表示数据库查询出错, -2表示用户不存在，返回0表示查询结果不一致。返回1表示一致
	user, isSuccess := logic.Signin(data)

	if isSuccess == -2 {
		code := UserNotExist
		ResponseError(c, code)
		return
	}

	if isSuccess == -1 {
		code := CodeServerBusy
		ResponseError(c, code)
		return
	}

	if isSuccess == 0 {
		code := LoginFaild
		ResponseError(c, code)
		return
	}

	// 登录成功
	ResponseSuccessed(c, gin.H{
		"user_id":       user.User_id,
		"user_name":     user.Username,
		"access_token":  user.AccessToken,
		"refresh_token": user.RefreshToken,
	})
}

// 对外提供RefreshToken的接口
func RefreshTokenHandler(c *gin.Context) {
	// 获取前端传来的RefreshToken
	rt := c.Query("refresh_token")

	// 验证RefreshToken是否合法
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" { // 如果是空的
		ResponseErrorWithMsg(c, InvalidToken, "请求头缺少Auth Token")
		c.Abort()
		return
	}

	// 获取authHeader后面的accessToken内容
	parts := strings.SplitN(authHeader, " ", 2)

	// 验证是否按规定传输
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		ResponseErrorWithMsg(c, InvalidToken, "Token格式不对")
		c.Abort()
		return
	}

	aToken, rToken, err := jwt.RefreshToken(parts[1], rt)
	fmt.Println(err)
	c.JSON(http.StatusOK, gin.H{
		"access_token":  aToken,
		"refresh_token": rToken,
	})
}
