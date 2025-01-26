package webSocket

import (
	"bluebell/models"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var (
	webSocketMap sync.Map // 用来存储用户的连接
	onlineCount  int32    // 在线人数
)

// 用来升级http请求为websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 发送消息
func sendMessage(user, message string) {

	webSocketMap.Range(func(key, value interface{}) bool {
		// 跳过发送给自己
		if key.(string) != user {
			log.Println(user, key.(string))
			if conn, ok := webSocketMap.Load(key.(string)); ok { // 发送给其他用户
				if err := conn.(*websocket.Conn).WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					log.Printf("Error sending message to %s: %s\n", user, err)
				}
			}
		}

		return true
	})

}

// 广播消息
func broadcastMessage(message string) {
	webSocketMap.Range(func(key, value interface{}) bool {

		go func(conn *websocket.Conn) {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				log.Printf("Error broadcasting message: %s\n", err)
			}
		}(value.(*websocket.Conn))
		return true
	})
}

// 广播用户列表
func broadcastUserEvent(user string, event string) {
	currentUsers := getCurrentUserList()

	broadcastMessage(currentUsers)
}

// 关闭某个用户的连接
func closeAndDeleteConnection(user string, conn *websocket.Conn) {
	err := conn.Close()
	if err != nil {
		log.Println("Error closing connection: ", err)
	}
	webSocketMap.Delete(user)
	reduceOnlineCount()
	// 从 firstIn.UserName 和 firstIn.UserInfo 中删除用户信息
	broadcastUserEvent(user, "离开了")
}

func addOnlineCount() int32 {
	return atomic.AddInt32(&onlineCount, 1)
}

func getOnlineCount() int32 {
	return atomic.LoadInt32(&onlineCount)
}

func reduceOnlineCount() int32 {
	return atomic.AddInt32(&onlineCount, -1)
}

func getCurrentUserList() string {
	var firstIn = models.FirstIn{
		Type: "userList",
	}
	firstIn.UserName = make([]string, 0)
	webSocketMap.Range(func(key, value interface{}) bool {
		firstIn.UserName = append(firstIn.UserName, key.(string))
		return true
	})
	userList, _ := json.Marshal(firstIn)
	return string(userList)
}
