package models

import "time"

type Community struct {
	CommunityId   uint64 `json:"community_id" db:"community_id"`
	CommunityName string `json:"community_name" db:"community_name"`
}

func (c *Community) TableName() string {
	return "community"
}

type CommunityDetail struct {
	CommunityID   uint64    `json:"community_id" db:"community_id"`
	CommunityName string    `json:"community_name" db:"community_name"`
	Introduction  string    `json:"introduction,omitempty" db:"introduction"` // omitempty 当Introduction为空时不展示
	CreateTime    time.Time `json:"create_time" db:"create_time"`
}

func (c *CommunityDetail) TableName() string {
	return "community"
}
