package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

const ContextUserIDKey = "userID"

var (
	ErrorUserNotLogin = errors.New("用户未登录")
)

/*
* 从路由中获取分页参数
 */
func getPageInfo(c *gin.Context) (int64, int64) {
	pageStr := c.Query("page") // 获取页数
	SizeStr := c.Query("size") // 获取每页大小

	var (
		page     int64
		pageSize int64
		err      error
	)
	page, err = strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		page = 1 // 默认第一页
	}

	pageSize, err = strconv.ParseInt(SizeStr, 10, 64)
	if err != nil {
		pageSize = 10 // 默认每页10条
	}

	return page, pageSize
}

// getCurrentUserId 从上下文中获取用户ID

func getCurrentUserId(c *gin.Context) (userId uint64, err error) {
	// 在进行Token验证的时候就将已经将用户ID存入上下文中
	userIdAny, ok := c.Get(ContextUserIDKey)
	if !ok {
		err = ErrorUserNotLogin
		return
	}

	userId, ok = userIdAny.(uint64)
	if !ok {
		err = ErrorUserNotLogin
		return
	}

	return
}
