package main

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/logic"
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
		return
	}

	redis.Init()

	if redis.RedisPool == nil {
		fmt.Println("init redis failed")
		return
	}

	rc := redis.RedisClient{
		PostService: mysql.NewPostService(),
	}

	r := gin.Default()

	routers.InitRouters(r, cfg)

	go logic.TimingtoStoreVotes()

	go rc.ListenForKeyspaceNotifications()

	r.Run()
}
