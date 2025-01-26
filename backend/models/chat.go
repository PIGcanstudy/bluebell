package models

type Message struct {
	UserName string `json:"userName"`
	Content  string `json:"Content"`
	Type     string `json:"type"`
}

type FirstIn struct {
	UserName []string `json:"user"` // 存储多个用户
	Type     string   `json:"type"` // 类型
}
