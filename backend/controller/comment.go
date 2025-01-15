package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CommentHandler(c *gin.Context) {
	var comment models.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		fmt.Println(err)
		ResponseError(c, InvalidParams)
		return
	}

	// 生成评论的ID
	commentId := snowflake.GetID()

	// 获取发起评论的人的ID
	authorId, err := getCurrentUserId(c)
	if err != nil {
		fmt.Println("GetCurrentUserID() failed ", err)
		ResponseErrorWithMsg(c, NotLogin, NotLogin.Msg())
		return
	}
	comment.CommentId = uint64(commentId)
	comment.AuthorId = authorId
	comment.CreateTime = time.Now()

	// 创建帖子
	if err := logic.CreateComment(&comment); err != nil {
		fmt.Println("CreateComment() failed ", err)
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccessed(c, nil)
}

// 获取有关文章的所有评论
func CommentListHandler(c *gin.Context) {
	// 获取文章的ID
	postId, err := strconv.ParseUint(c.Query("post_id"), 10, 64)
	if err != nil {
		fmt.Println("ParseUint() failed ", err)
		ResponseError(c, CodeServerBusy)
		return
	}
	var comments []models.Comment

	if err := logic.GetCommentList(postId, &comments); err != nil {
		fmt.Println("GetCommentList() failed ", err)
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccessed(c, comments)
}
