package main

import (
	"bluebell/dao/mysql"
	"bluebell/routers"
	"fmt"

	"github.com/gin-gonic/gin"
	"gopkg.in/ini.v1"
)

func main() {
	cfg, err := ini.Load("./conf/config.ini")

	if err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}

	// 初始化mysql
	if err = mysql.Init(cfg); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
	}

	r := gin.Default()

	routers.InitRouters(r)

	r.Run()
}
