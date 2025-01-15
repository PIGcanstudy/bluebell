package models

import "time"

type Comment struct {
	PostId     uint64    `json:"post_id" db:"post_id"`         // 文章的ID
	CommentId  uint64    `json:"comment_id" db:"comment_id"`   // 评论的ID
	AuthorId   uint64    `json:"author_id" db:"author_id"`     // 发起评论的是谁
	Content    string    `json:"content" db:"content"`         // 评论的内容
	CreateTime time.Time `json:"create_time" db:"create_time"` // 评论的创建时间
}
