package controller

import (
	"bluebell/logic"
	"bluebell/models"

	"github.com/gin-gonic/gin"
)

// 谁给哪个文章投了什么票
func VoteHandler(c *gin.Context) {
	vote := new(models.VoteDataForm)
	// 这里就可以知道投给哪篇文章，投的是什么类型的
	if err := c.ShouldBindJSON(vote); err != nil {
		ResponseErrorWithMsg(c, InvalidParams, InvalidParams.Msg())
		return
	}

	// 获取由谁发起的投票
	userId, err := getCurrentUserId(c)
	if err != nil {
		ResponseErrorWithMsg(c, NotLogin, NotLogin.Msg())
		return
	}

	if err := logic.VoteForPost(userId, vote); err != nil {
		if err.Error() == "已经投过票了，不能重复投票" {
			ResponseSuccessed(c, VoteRepated.Msg())
			return
		}
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccessed(c, nil)
}
