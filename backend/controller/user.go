package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"fmt"
	"net/http"

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
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  code.Msg(),
			"Data": nil,
		})
		return
	}

	// 验证确认密码和密码是否一致
	if data.Password != data.Confirm {
		code := PasswordNotMatch
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  code.Msg(),
			"Data": nil,
		})
		return
	}

	//  开始处理业务注册逻辑
	if isSuccess := logic.Signup(data); !isSuccess {
		code := UserExist
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  code.Msg(),
			"Data": nil,
		})
		return
	}

	// 已经注册完毕了
	code := ResponseSuccess
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  code.Msg(),
		"Data": "注册成功",
	})
}

func SigninHandler(c *gin.Context) {
	// 获取前端传来的参数
	var data models.LoginForm

	if err := c.ShouldBindJSON(&data); err != nil {
		fmt.Println("登录时参数错误", err)
		return
	}

	// 返回-1表示数据库查询出错, -2表示用户不存在，返回0表示查询结果不一致。返回1表示一致
	isSuccess := logic.Signin(data)

	if isSuccess == -2 {
		code := UserNotExist
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  code.Msg(),
			"Data": nil,
		})
		return
	}

	if isSuccess == -1 {
		code := CodeServerBusy
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  code.Msg(),
			"Data": nil,
		})
		return
	}

	if isSuccess == 0 {
		code := LoginFaild
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  code.Msg(),
			"Data": nil,
		})
		return
	}

	// 登录成功
	code := ResponseSuccess
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  code.Msg(),
		"Data": "登录成功",
	})

}
