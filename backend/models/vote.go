package models

import (
	"encoding/json"
	"errors"
)

type VoteDataForm struct {
	PostId    uint64 `json:"post_id" binding:"required"`   // 投给谁
	Direction *int8  `json:"direction" binding:"required"` // 方向，1为赞，-1为踩，0为取消
}

// UnmarshalJSON 为VoteDataForm类型实现自定义的UnmarshalJSON方法
func (v *VoteDataForm) UnmarshalJSON(data []byte) (err error) {
	required := struct {
		PostID    uint64 `json:"post_id"`
		Direction *int8  `json:"direction"`
	}{}
	err = json.Unmarshal(data, &required)
	if err != nil {
		return
	} else if required.PostID == 0 {
		err = errors.New("缺少必填字段post_id")
	} else if required.Direction == nil {
		err = errors.New("缺少必填字段direction")
	} else {
		v.PostId = required.PostID
		v.Direction = required.Direction
	}
	return
}

type VoteDataScore struct {
	PostId uint64 `json:"post_id"`
	Score  int64  `json:"score"`
}
