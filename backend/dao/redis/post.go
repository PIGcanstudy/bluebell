package redis

import (
	"bluebell/models"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	OneWeekInSeconds         = 7 * 24 * 3600
	VoteScore        float64 = 432 // 每一票的值432分
	PostPerAge               = 20
)

func CreatePost(postID, userID uint64, title, summary string, CommunityID uint64) (err error) {
	// 得到创建时间
	now := float64(time.Now().Unix())
	// 存入redis时候的key值
	votedKey := KeyPostVotedZSetPrefix + strconv.Itoa(int(postID))
	communityKey := KeyCommunityPostSetPrefix + strconv.Itoa(int(CommunityID))

	postInfo := map[string]interface{}{
		"title":    title,
		"summary":  summary,
		"post:id":  postID,
		"user:id":  userID,
		"time":     now,
		"likes":    1,
		"unlikes":  0,
		"comments": 0,
	}

	client := RedisPool.Get()
	defer client.Close()

	// 标记一个事务的开始
	client.Send("MULTI")

	// 将votedKey的key与value加1
	client.Send("ZADD", votedKey, 1, userID) // 作者默认投自己一票

	// 设置过期时间
	client.Send("Expire", votedKey, OneWeekInSeconds)

	// 构建 HMSET 的参数已经投过票了，不能重复投票
	args := []interface{}{KeyPostInfoHashPrefix + strconv.Itoa(int(postID))}
	for field, value := range postInfo {
		args = append(args, field, value)
	}

	// 发送 HMSET 命令 创建哈希表
	client.Send("HMSET", args...)

	// 因为作者投了自己一票，而一票的分数是432
	client.Send("ZADD", KeyPostScoreZSet, now+VoteScore, postID) // 添加到分数的ZSet

	client.Send("ZADD", KeyPostTimeZSet, now, postID)

	// 发布社区帖子的set
	client.Send("SADD", communityKey, postID) // 添加到对应版块  把帖子添加到社区的set
	_, err = client.Do("EXEC")
	return
}

// 查出所有文章的PostId的集合
func getIDsFormKey(key string, page, pageSize int64) ([]string, error) {
	client := RedisPool.Get()
	defer client.Close()
	start := (page - 1) * pageSize
	end := start + pageSize - 1
	// 3.ZREVRANGE 按照分数从大到小的顺序查询指定数量的元素
	rt, err := client.Do("ZREVRANGE", key, start, end)
	if err != nil {
		return nil, err
	}

	data, ok := rt.([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert redis response to []string")
	}

	results := make([]string, len(data))
	for i, v := range data {
		results[i] = string(v.([]byte))
	}

	return results, nil
}

func GetPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	// 从redis中获取id
	// 根据用户请求中携带的order参数确定要查询的redis key
	key := KeyPostTimeZSet
	if p.Order == models.OrderScoreDesc {
		key = KeyPostScoreZSet
	}

	// 确定查询的索引起始点
	return getIDsFormKey(key, p.Page, p.PageSize)
}

// 从key中分页取出指定页码上的帖子id
func GetPost(order string, page int64) []map[string]string {
	client := RedisPool.Get()
	defer client.Close()

	key := KeyPostScoreZSet
	if order == models.OrderTimeDesc {
		key = KeyPostTimeZSet
	}

	start := (page - 1) * PostPerAge
	end := start + PostPerAge - 1
	// 取出指定页码上的所有id
	idss, err := client.Do("ZREVRANGE", key, start, end)
	if err != nil {
		fmt.Println("failed to get post ids from redis:", err)
		return nil
	}

	ids, err := redis.Values(idss, err)
	if err != nil {
		fmt.Println("failed to convert redis response to []interface{}:", err)
		return nil
	}

	postList := make([]map[string]string, len(ids))

	// 遍历指定页码上的所有id
	for _, id := range ids {
		postdata, err := client.Do("HGETALL", KeyPostInfoHashPrefix+string(id.([]byte)))
		if err != nil {
			fmt.Println("failed to get post data from redis:", err)
			return nil
		}
		postData, _ := redis.StringMap(postdata, err)

		postData["id"] = id.(string)
		postList = append(postList, postData)

	}
	return postList
}

// 分社区根据发帖时间或者分数取出分页的帖子id
func GetCommunityPost(communityName, orderKey string, page int64) []map[string]string {
	client := RedisPool.Get()
	defer client.Close()

	key := orderKey + communityName // 创建缓存键

	isexist, err := client.Do("Exist", key)
	isExist, _ := redis.Bool(isexist, err)

	if !isExist {
		// 不存在则创建缓存
		_, err := client.Do("ZInterStore", key, 2, KeyCommunityPostSetPrefix+communityName, orderKey, "AGGREGATE", "MAX")
		if err != nil {
			fmt.Println("failed to create cache:", err)
			return nil
		}
		client.Do("EXPIRE", key, 60*time.Second)
	}
	return GetPost(orderKey, page)
}

// Reddit Hot rank algorithms
// from https://github.com/reddit-archive/reddit/blob/master/r2/r2/lib/db/_sorts.pyx
func Hot(ups, downs int, date time.Time) float64 {
	// 对应votea - voteb
	s := float64(ups - downs)
	// math.Max(math.Abs(s), 1)) 对应 if |s| < 1: s = 1 else : s
	order := math.Log10(math.Max(math.Abs(s), 1))
	var sign float64
	if s > 0 {
		sign = 1
	} else if s == 0 {
		sign = 0
	} else {
		sign = -1
	}
	// 对应ts = ta - tb
	seconds := float64(date.Second() - 1577808000)
	// 取整
	return math.Round(sign*order + seconds/43200)
}

// 从rediszset中获取所有超过一周的帖子的投票记录
func GetAllVoteRecords() (data []models.PostVoteDetail, err error) {
	now := time.Now().Unix()
	keyPattern := KeyPostInfoHashPrefix + "*"
	client := RedisPool.Get()
	defer client.Close()

	var cursor int64

	for {
		// 使用SCAN命令扫描匹配的key
		reply, err := redis.Values(client.Do("SCAN", cursor, "MATCH", keyPattern))
		if err != nil {
			fmt.Println("failed to scan keys:", err)
			return nil, err
		}

		// 解析返回的值
		cursor, _ = redis.Int64(reply[0], nil)  // 获取新的游标
		keys, _ := redis.Strings(reply[1], nil) // 获取匹配的键

		// 遍历每个key，获取Likes和Unlikes以及time字段
		for _, key := range keys {
			likes, err := redis.Int64(client.Do("HGET", key, "likes")) // 获取Likes字段
			if err != nil {
				fmt.Println("failed to get Likes for key:", key, err)
				continue
			}

			unlikes, err := redis.Int64(client.Do("HGET", key, "unlikes")) // 获取Unlikes字段
			if err != nil {
				fmt.Println("failed to get Unlikes for key:", key, err)
				continue
			}

			timeStamp, err := redis.Int64(client.Do("HGET", key, "time")) // 获取time字段
			if err != nil {
				fmt.Println("failed to get time for key:", key, err)
				continue
			}

			// 判断timeStamp是否过了7天
			if now-timeStamp > OneWeekInSeconds {
				continue
			}

			// 截取postid
			postid, err := strconv.ParseUint(key[len(KeyPostInfoHashPrefix):], 10, 64)
			if err != nil {
				fmt.Println("failed to parse postid from key:", key, err)
				continue
			}

			// 将结果添加到切片中
			postVoteDetail := models.PostVoteDetail{
				PostId:  postid,
				Likes:   likes,
				Unlikes: unlikes,
			}

			data = append(data, postVoteDetail)
		}

		if cursor == 0 {
			break
		}
	}
	return
}
