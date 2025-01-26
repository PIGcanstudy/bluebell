package webSocket

import (
	"bluebell/models"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ChatHandler(c *gin.Context) {
	// 先查询是否有 user 参数
	user := c.Query("user")
	if user == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名参数不能为空"})
		return
	}
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败：", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "WebSocket 升级失败"})
		return
	}

	defer closeAndDeleteConnection(user, ws)
	webSocketMap.Store(user, ws)
	addOnlineCount()
	// 广播某个用户加入了聊天室
	broadcastUserEvent(user, "加入了")
	log.Printf("用户连接：%s，当前在线人数：%d\n", user, getOnlineCount())

	for { // 每个用户链接都会从websocket中循环读数据，如果没有数据到来，会阻塞在这里
		_, p, err := ws.ReadMessage()
		if err != nil {
			break
		}
		var msg models.Message
		if err := json.Unmarshal(p, &msg); err != nil {
			sendMessage(user, "消息格式错误："+err.Error())
			continue
		}
		message, _ := json.Marshal(msg)
		sendMessage(user, string(message))
	}
}
