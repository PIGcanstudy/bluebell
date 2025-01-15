package controller

import (
	"bluebell/logic"
	"bluebell/models"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

/*
* 功能：分页展示帖子列表
 */
func GetPostListHandler(c *gin.Context) {
	// 获取分页参数
	page, pageSize := getPageInfo(c)

	// 根据分页参数查询数据
	data, err := logic.GetPostList(page, pageSize)
	if err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccessed(c, data)
}

// 通过ID查询帖子详情
func GetPostByIdHandler(c *gin.Context) {
	// 从路径参数中获取帖子ID
	postIdStr := c.Param("id")

	postId, err := strconv.ParseUint(postIdStr, 10, 64)
	if err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	// 根据id取出id对应的帖子详情信息
	post, err := logic.GetPostById(postId)
	if err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccessed(c, post)
}

// 创建帖子(本质就是创建一个post)
func CreatePostHandler(c *gin.Context) {
	var post models.Post

	// 解析参数到Post结构体中
	if err := c.ShouldBindJSON(&post); err != nil {
		fmt.Println("create post failed, invalid params", err.Error())
		ResponseErrorWithMsg(c, InvalidParams, err.Error())
		return
	}

	// 获取当前作者作者Id
	userId, err := getCurrentUserId(c)

	if err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	post.AuthorId = userId

	// 调用逻辑层的创建帖子逻辑
	if err = logic.CreatePost(&post); err != nil {
		fmt.Println("create post failed, server busy", err.Error())
		ResponseError(c, CodeServerBusy)
		return
	}

	// 成功就放回响应
	ResponseSuccessed(c, post)
}

func GetPostListSortedHandler(c *gin.Context) {
	// 获取分页查询参数
	p := &models.ParamPostList{
		Page:     1,
		PageSize: 10,
		Order:    models.OrderTimeDesc,
	}

	//c.ShouldBind() 根据请求的数据类型选择相应的方法去获取数据
	//c.ShouldBindJSON() 如果请求中携带的是json格式的数据，才能用这个方法获取到数据
	if err := c.ShouldBindQuery(p); err != nil {
		fmt.Println("get post list sorted failed, invalid params", err.Error())
		ResponseError(c, InvalidParams)
		return
	}

	// 从数据库中获取数据
	data, err := logic.GetPostListSorted(p)
	if err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccessed(c, data)
}
