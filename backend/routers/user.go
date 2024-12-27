package routers

import (
	"bluebell/controller"

	"github.com/gin-gonic/gin"
)

func InitRouters(r *gin.Engine) {
	v := r.Group("/user")

	v.POST("/signup", controller.SignupHandler)
	v.POST("/signin", controller.SigninHandler)
}
