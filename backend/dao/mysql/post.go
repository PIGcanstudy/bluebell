package mysql

import (
	"bluebell/dao/interfaces"
	"bluebell/dao/redis"
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

func DeletePostById(postId uint64) (err error) {
	// 开启事务
	DB.Begin()

	// 删除前先把删除的内容给保存一份
	post, err := redis.GetPostById(postId)

	// 调用redis的相关删除
	if err = redis.DeletePostById(postId); err != nil {
		// redis删除失败, 由于没有执行任何操作直接返回
		DB.Rollback()
		return
	}

	// redis删除成功, 删除mysql中的数据
	if errs := DB.Table("post").Delete("*", "post_id = ?", postId).Error; errs != nil {
		// mysql删除失败, 恢复redis的数据
		fmt.Println("DeletePostById error: ", errs)
		if errss := redis.RestorePostById(postId, post); errss != nil {
			fmt.Println("DeletePostById error: ", errss)
			fmt.Println("开始重试操作")
			// 恢复redis失败，使用重试机制确保成功，开一个协程来进行
		}
		DB.Rollback()
		return errs
	}

	// 两个都删除成功了，就提交事务
	DB.Commit()
	return nil
}

func UpdatePostContent(postId uint64, postContent models.PostContent) error {
	rt := DB.Table("post").Where("post_id = ?", postId).Updates(map[string]interface{}{
		"content":      postContent.Content,
		"title":        postContent.Title,
		"community_id": postContent.Community_id,
	})
	if rt.Error != nil {
		return rt.Error
	}
	return nil

}

type PostServiceImpl struct {
}

func (m PostServiceImpl) InsertPostVoteRecord(postId, userId uint64, score int) error {

	var str string

	if score == 432 {
		str = "like"
	} else if score == -432 {
		str = "unlike"
	}
	rt := DB.Table("post_votes").Create(&models.VoteDataScore{
		PostId:   postId,
		UserId:   userId,
		VoteType: str,
	})
	if rt.Error != nil {
		fmt.Println("InsertPostVoteRecord error: ", rt.Error)
		return rt.Error
	}

	// 插入成功
	return nil
}

func NewPostService() interfaces.PostService {
	return &PostServiceImpl{}
}
