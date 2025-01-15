package controller

import (
	"bluebell/logic"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 获取社区列表
func GetCommunityListHandler(c *gin.Context) {
	communityList, err := logic.GetCommunityList()
	if err != nil {
		zap.L().Error("logic.GetCommunityList() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy) // 不轻易把服务端报错暴露给外面
		return
	}

	ResponseSuccessed(c, communityList)
}

func SearchCommunityByIdHandler(c *gin.Context) {
	// 首先从查询参数中获取社区ID
	id := c.Param("id")
	fmt.Println(id)
	if id == "" {
		ResponseError(c, InvalidParams)
		return
	}

	ID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		ResponseError(c, InvalidParams)
		return
	}

	detail, err := logic.SearchCommunityById(ID)

	if err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccessed(c, detail)
}
