package models

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	OrderTimeDesc  = "time"
	OrderScoreDesc = "score"
)

// 帖子post结构体
type Post struct {
	PostID      uint64    `json:"post_id" db:"post_id"`
	AuthorId    uint64    `json:"author_id" db:"author_id"`
	CommunityId uint64    `json:"community_id" db:"community_id"`
	Likes       int64     `json:"likes" db:"likes"`
	Dislikes    int64     `json:"dislikes" db:"dislikes"`
	Status      int32     `json:"status" db:"status"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	CreateTime  time.Time `json:"create_time" db:"create_time"`
}

func (p *Post) TableName() string {
	return "post"
}

// UnmarshalJSON 为Post类型实现自定义的UnmarshalJSON方法
func (p *Post) UnmarshalJSON(data []byte) (err error) {
	required := struct {
		Title       string `json:"title" db:"title"`
		Content     string `json:"content" db:"content"`
		CommunityID int64  `json:"community_id" db:"community_id"`
	}{}
	err = json.Unmarshal(data, &required)
	if err != nil {
		return
	} else if len(required.Title) == 0 {
		err = errors.New("帖子标题不能为空")
	} else if len(required.Content) == 0 {
		err = errors.New("帖子内容不能为空")
	} else if required.CommunityID == 0 {
		err = errors.New("未指定版块")
	} else {
		p.Title = required.Title
		p.Content = required.Content
		p.CommunityId = uint64(required.CommunityID)
	}
	return
}

// 帖子返回的详情结构体
type PostDetail struct {
	Post                               // 嵌入帖子结构体
	CommunityDetail `json:"community"` // 嵌入版块详情结构体
	AuthorName      string             `json:"author_name"`
	VoteNum         int64              `json:"vote_num"` // 帖子的点赞数
}

type ParamPostList struct {
	CommunityId uint64 `json:"community_id" form:"community_id"`
	Page        int64  `json:"page" form:"page"`
	PageSize    int64  `json:"page_size" form:"page_size"`
	Order       string `json:"order" form:"order"`
}

type PostVoteDetail struct {
	PostId  uint64 `json:"post_id"`
	Likes   int64  `json:"like"`
	Unlikes int64  `json:"unlike"`
}

type PostContent struct {
	Content      string `json:"content" db:"content" redis:"summary"`
	Title        string `json:"title" db:"title" redis:"title"`
	Community_id int64  `json:"community_id" db:"community_id" redis:"community_id"`
}
