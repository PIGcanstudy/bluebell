package routers

import (
	"bluebell/controller"
	"bluebell/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gopkg.in/ini.v1"
)

func InitRouters(r *gin.Engine, cfg *ini.File) {

	//设置启动模式
	gin.SetMode(cfg.Section("system").Key("env").String())

	// 全局设置响应头
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://127.0.0.1:5173")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization")
		c.Next()
	})
	//宕机时恢复
	r.Use(gin.Recovery())

	r.StaticFS(cfg.Section("upload").Key("path").String(), http.Dir(cfg.Section("upload").Key("path").String()))

	// swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 无效路由报404
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "404 not found",
		})
	})

	v := r.Group("/user")

	v.POST("/signup", controller.SignupHandler)             // 注册业务
	v.POST("/signin", controller.SigninHandler)             // 登录业务
	v.GET("/refresh_token", controller.RefreshTokenHandler) // 刷新token业务

	// 获取分类社区列表
	v.GET("/community/list", controller.GetCommunityListHandler)

	// 根据ID查找社区详情
	v.GET("/community/:id", controller.SearchCommunityByIdHandler)

	// 根据时间或者分数排序分页展示帖子列表
	v.GET("/post/list/sorted", controller.GetPostListSortedHandler)

	// 分页展示帖子列表
	v.GET("/post/list", controller.GetPostListHandler)

	// 查询帖子详情
	v.GET("/post/:id", controller.GetPostByIdHandler)

	v.Use(middlewares.JWTAuthMiddleware())
	{
		// 创建帖子
		v.POST("/post", controller.CreatePostHandler)
		// 更新帖子
		v.POST("/UpdatePost", controller.UpdatePostHandler)
		// 删除帖子
		v.POST("/DeletePost", controller.DeletePostHandler)

		// 投票
		v.POST("/vote", controller.VoteHandler)

		// 评论
		v.POST("/comment", controller.CommentHandler)

		// 退出登录
		v.POST("/signout", controller.SignoutHandler)

		v.GET("/comment", controller.CommentListHandler)

		// 用户之间的私聊
		v.GET("/chat", controller.ChatHandler)

		v.GET("/ping", func(c *gin.Context) {
			c.String(200, "pong")
		}) // 测试用
	}
}
