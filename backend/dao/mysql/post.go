package mysql

import (
	"bluebell/models"
	"fmt"
	"strconv"
)

func GetPostList(page int64, pageSize int64) (posts []models.Post, err error) {
	k := (page - 1) * pageSize
	rt := DB.Debug().Table("post").Select("post_id, title, content, author_id, community_id, create_time").Order("create_time desc").Limit((int(pageSize))).Offset(int(k)).Scan(&posts)
	if rt.Error != nil {
		fmt.Println("GetPostList error: ", rt.Error)
		return nil, rt.Error
	}

	return
}

func GetPostByIds(ids []string) (posts []models.Post, err error) {
	posts = make([]models.Post, len(ids))
	// 这种方式太耗时间了，应该尽量减少查询数据库的次数
	// for _, id := range ids {
	// 	Id, _ := strconv.ParseUint(id, 10, 64)
	// 	post, err := GetPostById(Id)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	posts = append(posts, post)
	// }

	// 改用下列这种方式, 使用In(语法)
	// 首先构造占位符
	uintIds := make([]uint64, len(ids))
	for i, id := range ids {
		Id, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			return nil, err
		}
		uintIds[i] = Id
	}

	if err := DB.Debug().Table("post").Where("post_id IN (?)", uintIds).Find(&posts).Error; err != nil {
		fmt.Println("GetPostByIds error: ", err)
		return nil, err
	}
	return
}

func GetPostById(postId uint64) (post models.Post, err error) {
	rt := DB.Debug().Where("post_id = ?", postId).First(&post)
	err = rt.Error
	if err != nil {
		fmt.Println("GetPostById error: ", err)
		return models.Post{}, err
	}

	return
}

func CreatePost(post *models.Post) (err error) {
	rt := DB.Table("post").Create(post)
	err = rt.Error
	return
}

func SaveVotesToMySQL(PostId uint64, likes int64, unlikes int64) (err error) {
	rt := DB.Table("post").Where("post_id = ?", PostId).Update("likes", likes).Update("dislikes", unlikes)
	return rt.Error
}
