package redis

import (
	"bluebell/dao/interfaces"
	"bluebell/models"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	OneWeekInSeconds         = 7 * 24 * 3600
	VoteScore        float64 = 432 // 每一票的值432分
	PostPerAge               = 20
)

type RedisClient struct {
	PostService interfaces.PostService
}

func RestorePostById(postId uint64, postdata map[string]interface{}) error {

	client := RedisPool.Get()
	defer client.Close()

	// 开始事务
	client.Send("MULTI")

	key_post_info := KeyPostInfoHashPrefix + strconv.Itoa(int(postId))
	key_post_score := KeyPostScoreZSet
	key_post_time := KeyPostTimeZSet
	key_post_voted := KeyPostVotedZSetPrefix + strconv.Itoa(int(postId))

	// 重置redis中存储的文章信息(是一个哈希结构)
	tmp_post_info := postdata[key_post_info]
	data_post_info, _ := redis.StringMap(tmp_post_info, nil)
	err := client.Send("HMSET", key_post_info, data_post_info)

	if err != nil {
		fmt.Println("failed to restore post info:", err)
		client.Send("DISCARD")
		return err
	}

	// 重置redis中的score字段
	post_score, err := redis.Values(postdata[key_post_score], nil)
	if err != nil {
		fmt.Println("查询有序集合失败:", err)
		client.Send("DISCARD")
		return err
	}

	for i := 0; i < len(post_score); i += 2 {
		value, err := redis.Uint64(post_score[i], nil)
		if err != nil {
			fmt.Println("无法转换 value 为 []byte")
			continue
		}

		score, err := redis.Int64(post_score[i+1], nil)
		if err != nil {
			fmt.Println("无法转换 score 为 int64:", err)
			continue
		}

		if err = client.Send("ZADD", key_post_score, score, value); err != nil {
			fmt.Println("failed to restore post score:", err)
			client.Send("DISCARD")
			return err
		}
	}

	// 重置redis中的time字段
	post_time, err := redis.Values(postdata[key_post_time], nil)
	if err != nil {
		fmt.Println("查询有序集合失败:", err)
		client.Send("DISCARD")
		return err
	}

	for i := 0; i < len(post_time); i += 2 {
		postId, err := redis.Uint64(post_time[i], nil)
		if err != nil {
			fmt.Println("无法转换 postId 为 uint64:", err)
			continue
		}

		score, err := redis.Int64(post_time[i+1], nil)
		if err != nil {
			fmt.Println("无法转换 score 为 int64:", err)
			continue
		}

		if err = client.Send("ZADD", key_post_time, score, postId); err != nil {
			fmt.Println("failed to restore post time:", err)
			client.Send("DISCARD")
			return err
		}
	}

	// 恢复redis中的投票字段
	post_voted, err := redis.Values(postdata[key_post_voted], nil)
	if err != nil {
		fmt.Println("查询有序集合失败:", err)
		client.Send("DISCARD")
		return err
	}

	for i := 0; i < len(post_voted); i += 2 {
		value, err := redis.Uint64(post_voted[i], nil)
		if err != nil {
			fmt.Println("无法转换 value 为 []byte")
			continue
		}

		score, err := redis.Int64(post_voted[i+1], nil)
		if err != nil {
			fmt.Println("无法转换 score 为 int64:", err)
			continue
		}

		if err = client.Send("ZADD", key_post_voted, score, value); err != nil {
			fmt.Println("failed to restore post voted:", err)
			client.Send("DISCARD")
			return err
		}
	}

	// 10 9 8 7 6 5 4 3 2 1 0
	return nil
}

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

	// 将votedKey的key与value加432
	client.Send("ZADD", votedKey, 432, userID) // 作者默认投自己一票

	// 设置过期时间
	client.Send("Expire", votedKey, OneWeekInSeconds)

	// 构建 HMSET 的参数已经投过票了，不能重复投票
	args := []interface{}{KeyPostInfoHashPrefix + strconv.Itoa(int(postID))}
	for field, value := range postInfo {
		args = append(args, field, value)
	}

	// 发送 HMSET 命令 创建哈希表
	client.Send("HMSET", args...)

	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)

	// 设置随机的过期时间，一定程度上缓解缓存雪崩问题
	client.Send("EXPIRE", KeyPostInfoHashPrefix+strconv.Itoa(int(postID)), OneWeekInSeconds+randGenerator.Intn(5001))

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

func GetPostById(postId uint64) (postData map[string]interface{}, err error) {
	client := RedisPool.Get()
	defer client.Close()

	postData = make(map[string]interface{})
	key_post_info := KeyPostInfoHashPrefix + strconv.Itoa(int(postId))
	key_post_score := KeyPostScoreZSet
	key_post_time := KeyPostTimeZSet
	key_post_voted := KeyPostVotedZSetPrefix + strconv.Itoa(int(postId))

	// 备份hash里的数据
	postdata, err := client.Do("HGETALL", key_post_info)
	if err != nil {
		fmt.Println("failed to get post data from redis:", err)
		return nil, err
	}

	postData[key_post_info] = postdata

	// 备份score和time字段
	postdata, err = client.Do("ZRANGE", key_post_score, 0, -1, "WITHSCORES")
	if err != nil {
		fmt.Println("failed to get post score from redis:", err)
		return nil, err
	}
	postData[key_post_score] = postdata

	postdata, err = client.Do("ZRANGE", key_post_time, 0, -1, "WITHSCORES")
	if err != nil {
		fmt.Println("failed to get post time from redis:", err)
		return nil, err
	}
	postData[key_post_time] = postdata

	// 备份投票字段
	postdata, err = client.Do("ZRANGE", key_post_voted, 0, -1, "WITHSCORES")
	if err != nil {
		fmt.Println("failed to get post voted from redis:", err)
		return nil, err
	}
	postData[key_post_voted] = postdata

	return postData, nil
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

func CheckPostIsExist(postId uint64) (isExist bool, err error) {
	client := RedisPool.Get()
	defer client.Close()

	key := KeyPostInfoHashPrefix + strconv.Itoa(int(postId))

	rt, err := client.Do("EXISTS", key)
	if err != nil {
		fmt.Println("failed to check post exist:", err)
		return false, err
	}

	isExist, _ = redis.Bool(rt, nil)

	return
}

func DeletePostById(postId uint64) error {
	client := RedisPool.Get()
	defer client.Close()

	// 开始事务
	client.Send("MULTI")

	key_post_info := KeyPostInfoHashPrefix + strconv.Itoa(int(postId))
	key_post_score := KeyPostScoreZSet
	key_post_time := KeyPostTimeZSet
	key_post_voted := KeyPostVotedZSetPrefix + strconv.Itoa(int(postId))

	// 删除与文章有关的key
	// 删除redis中存储的文章信息
	client.Send("DEL", key_post_info)

	// 删除redis中time与score字段
	client.Send("ZREM", key_post_score, postId)
	client.Send("ZREM", key_post_time, postId)

	// 删除投票字段
	client.Send("DEL", key_post_voted)

	// 执行事务
	_, err := client.Do("EXEC")

	if err != nil {
		fmt.Println("failed to delete post:", err)
		return err
	}

	return nil
}

func (r *RedisClient) ListenForKeyspaceNotifications() {
	client := RedisPool.Get()
	defer client.Close()

	// 订阅键过期事件频道
	psc := redis.PubSubConn{Conn: client}
	// 与Subscribe()不同，Subscribe()只能订阅一个频道，而PSubscribe()可以订阅多个匹配的频道
	err := psc.PSubscribe("__keyevent@0__:expired")
	if err != nil {
		fmt.Println("failed to subscribe to keyspace notifications:")
		return
	}

	fmt.Println("Subscribed to Redis keyspace notifications...")

	// 循环处理订阅到的消息
	for {
		switch msg := psc.Receive().(type) {
		case redis.Message:
			// 当键过期时，msg.Data为服务的键名
			expiredKey := string(msg.Data)

			// 从键名中获取postId
			postId, err := strconv.ParseUint(expiredKey[len(KeyPostInfoHashPrefix):], 10, 64)
			if err != nil {
				fmt.Println("failed to parse postId from key:", expiredKey, err)
				continue
			}

			fmt.Printf("Key %s expired\n", expiredKey)

			// 使用PERSIST取消键的过期时间
			_, err = client.Do("PERSIST", expiredKey)
			if err != nil {
				fmt.Println("failed to persist key:", err)
				continue
			}

			// 获取过期键的值
			value, err := redis.Values(client.Do("ZRANGE", expiredKey, 0, -1, "WITHSCORES"))
			if err != nil {
				fmt.Println("failed to get expired key value:", err)
				continue
			}

			// 将数据存入到mysql中
			for i := 0; i < len(value); i += 2 {
				userId, err := redis.Uint64(value[i], nil)
				if err != nil {
					fmt.Println("failed to parse postId from redis:", err)
					continue
				}

				score, err := redis.Int(value[i+1], nil)
				if err != nil {
					fmt.Println("failed to parse score from redis:", err)
					continue
				}

				// 存入mysql中
				if err = r.PostService.InsertPostVoteRecord(postId, userId, score); err != nil {
					fmt.Println("failed to insert post vote record:", err)
					continue
				}
			}
			// 删除redis中的投票记录
			if _, err = client.Do("DEL", expiredKey); err != nil {
				fmt.Println("failed to delete post vote record:", err)
				continue
			}
		case redis.Subscription:
			// 订阅成功
			fmt.Printf("Subscribed to: %s (kind: %s, count: %d)\n", msg.Channel, msg.Kind, msg.Count)
		case error:
			// 处理错误
			fmt.Println("Error: %v\n", msg)
			return
		}
	}
}

func UpdatePostContent(postId uint64, content models.PostContent, oldCommunityId uint64) error {
	client := RedisPool.Get()
	defer client.Close()

	key_post_info := KeyPostInfoHashPrefix + strconv.Itoa(int(postId))

	// 开始事务
	client.Send("MULTI")

	// 更新文章内容
	client.Send("HSET", key_post_info, "title", content.Title)
	client.Send("HSET", key_post_info, "summary", content.Content)

	key_post_community := KeyCommunityPostSetPrefix + strconv.Itoa(int(content.Community_id))
	key_post_oldCommunity := KeyCommunityPostSetPrefix + strconv.Itoa(int(oldCommunityId))

	// 更新文章社区
	client.Send("SREM", key_post_oldCommunity, postId)
	client.Send("SADD", key_post_community, postId)

	// 执行事务
	_, err := client.Do("EXEC")

	if err != nil {
		client.Send("DISCARD")
		return err
	}

	return nil
}
