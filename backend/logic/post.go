package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"fmt"
	"time"
)

func GetPostList(page int64, size int64) (data []models.PostDetail, err error) {
	postList, err := mysql.GetPostList(page, size)

	if err != nil {
		return nil, err
	}

	data = make([]models.PostDetail, 0)
	fmt.Println(len(data))

	for _, post := range postList {
		// 查找作者信息
		author, err := mysql.GetUserByID(post.AuthorId)
		if err != nil {
			continue
		}

		// 查找社区信息
		var community models.CommunityDetail
		err = mysql.GetCommunityById(post.CommunityId, &community)
		if err != nil {
			continue
		}
		// 往data中加入数据
		postdetail := models.PostDetail{
			Post:            post,
			AuthorName:      author.Username,
			CommunityDetail: community,
			VoteNum:         0,
		}

		data = append(data, postdetail)
	}
	return
}

func GetPostById(postId uint64) (data models.PostDetail, err error) {
	post, err := mysql.GetPostById(postId)
	if err != nil {
		return
	}

	// 查找作者信息
	author, err := mysql.GetUserByID(post.AuthorId)
	if err != nil {
		return
	}

	// 查找社区信息
	var community models.CommunityDetail
	err = mysql.GetCommunityById(post.CommunityId, &community)
	if err != nil {
		return
	}

	// 接口数据拼接
	data = models.PostDetail{
		Post:            post,
		AuthorName:      author.Username,
		CommunityDetail: community,
		VoteNum:         0,
	}

	return
}

// 创建帖子的逻辑
func CreatePost(post *models.Post) (err error) {
	// 先生成PostID
	postId := snowflake.GetID()

	post.PostID = uint64(postId)
	post.CreateTime = time.Now()

	// 创建帖子 保存到数据库中
	if err = mysql.CreatePost(post); err != nil {
		fmt.Println("创建帖子失败，其错误码是: ", err)
		return
	}

	// 查找社区信息
	var community models.CommunityDetail
	err = mysql.GetCommunityById(post.CommunityId, &community)
	if err != nil {
		fmt.Println("mysql.GetCommunityById() 查询社区信息失败 ", err)
		return
	}

	// 使用redis缓存帖子信息（这样可以提高用户加载效率）
	if err = redis.CreatePost(
		post.PostID,
		post.AuthorId,
		post.Title,
		TruncateByWords(post.Content, 120),
		community.CommunityID); err != nil {
		fmt.Println("redis 缓存帖子信息失败，其错误码是: ", err)
		return
	}

	return
}

// 排行版使用
func GetPostListSorted(p *models.ParamPostList) (data []*models.PostDetail, err error) {
	// 去redis查询id列表
	ids, err := redis.GetPostIDsInOrder(p)
	if err != nil {
		return
	}

	if len(ids) == 0 {
		fmt.Println("redis.GetPostIDsInOrder() 返回的id列表为空")
		return
	}

	// 提前查询好每篇帖子的投票数(顺序与ids的顺序一致)
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		fmt.Println("redis.GetPostVoteData() 查询投票数失败 ", err)
		return
	}

	// 根据id去数据库查询帖子的详情信息
	// 返回的数据还要按照给定的id的顺序返回
	posts, err := mysql.GetPostByIds(ids)
	if err != nil {
		return
	}

	for idx, post := range posts {
		// 将帖子的作者的名字
		author, err := mysql.GetUserByID(post.AuthorId)
		if err != nil {
			fmt.Println("mysql.GetUserByID() 查询作者信息失败 ", err)
			continue
		}
		// 更具社区id查询社区详细信息
		var community models.CommunityDetail
		err = mysql.GetCommunityById(post.CommunityId, &community)
		if err != nil {
			fmt.Println("mysql.GetCommunityById() 查询社区信息失败 ", err)
			continue
		}
		// 接口数据拼接
		postdetail := &models.PostDetail{
			Post:            post,
			CommunityDetail: community,
			AuthorName:      author.Username,
			VoteNum:         voteData[idx],
		}
		data = append(data, postdetail)
	}
	return
}

func TimingtoStoreVotes() {
	// 定时器(每一小时检查一次)
	ticker := time.NewTicker(time.Hour)
	c := ticker.C
	for {
		select {
		case <-c:
			// 定时器到期，检查并存储投票数据
			checkAndStoreVotes()
		}
	}
}

func checkAndStoreVotes() {
	// 获取所有的投票记录(通过Redis查询postInfo中的投票数据)
	voteRecords, err := redis.GetAllVoteRecords()

	if err != nil {
		fmt.Println("redis.GetAllVoteRecords() 查询投票记录失败 ", err)
		return
	}

	for _, record := range voteRecords {
		// 将数据存入MySQL
		err := mysql.SaveVotesToMySQL(record.PostId, record.Likes, record.Unlikes)
		if err != nil {
			fmt.Println("mysql.SaveVotesToMySQL() 存储投票数据失败 ", err)
			// 处理存储错误
			continue
		}
	}
}
