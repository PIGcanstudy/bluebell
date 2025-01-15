package mysql

import (
	"bluebell/models"
	"fmt"
)

func GetCommunityList(communityList *[]models.Community) {
	rt := DB.Select("community_id, community_name").Find(&communityList)
	if rt.Error != nil {
		fmt.Printf("Error: %v", rt.Error)
		return
	}
}

func GetCommunityById(communityId uint64, community *models.CommunityDetail) error {
	rt := DB.Table("community").Where("community_id = ?", communityId).Scan(&community)
	if rt.Error != nil {
		fmt.Printf("Error: %v", rt.Error)
		return rt.Error
	}
	return nil
}
