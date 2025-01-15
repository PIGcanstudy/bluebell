package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"errors"
	"fmt"
)

func GetCommunityList() (communityList []models.Community, err error) {
	mysql.GetCommunityList(&communityList)
	if communityList == nil {
		err = errors.New("community list is nil")
	}
	return
}

func SearchCommunityById(id uint64) (communityDetail models.CommunityDetail, err error) {
	mysql.GetCommunityById(id, &communityDetail)
	if communityDetail == (models.CommunityDetail{}) {
		fmt.Println(2)
		err = errors.New("community is nil")
	}
	return
}
