package middlewares

import (
	"bluebell/controller"
	"bluebell/pkg/jwt"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware 基于JWT的认证中间件
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URL
		// 这里假设Token放在Header的Authorization中，并使用Bearer开头
		// 这里的具体实现方式要依据你的实际业务情况决定

		// 首先从请求头中读出Authorization字段
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" { // 如果是空的
			controller.ResponseErrorWithMsg(c, controller.InvalidToken, "请求头缺少Auth Token")
			c.Abort()
			return
		}
		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		// 看token格式是否正确
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			controller.ResponseErrorWithMsg(c, controller.InvalidToken, "Token格式不对")
			c.Abort()
			return
		}
		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			fmt.Println(err)
			controller.ResponseError(c, controller.InvalidToken)
			c.Abort()
			return
		}
		// 将当前请求的userID信息保存到请求的上下文c上
		c.Set(controller.ContextUserIDKey, mc.UserId)
		c.Next() // 后续的处理函数可以用过c.Get(ContextUserIDKey)来获取当前请求的用户信息
	}
}
