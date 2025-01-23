package redis

import (
	"bluebell/models"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

func VoteForPost(userId uint64, post *models.VoteDataForm) error {
	direction := post.Direction
	postId := post.PostId

	v := float64(*direction) * VoteScore

	client := RedisPool.Get()
	defer client.Close()

	// 首先先判断投票限制
	// 去Redis中取得帖子的发表时间
	PostTime, err := client.Do("ZSCORE", KeyPostTimeZSet, postId)
	if err != nil {
		fmt.Println("ZSCORE error:", err)
		return err
	}

	PostTimeInt, _ := redis.Int64(PostTime, err)

	// 超过一周的发表时间，不能再投票
	if time.Now().Unix()-PostTimeInt > OneWeekInSeconds {

		err := errors.New("超过一周的发表时间，不能再投票")
		return err
	}

	// 开始更新帖子的分数
	// 首先先判断是否投过票，查询当前用户给当前帖子的投票情况
	key := KeyPostVotedZSetPrefix + fmt.Sprintf("%d", postId)
	ov, _ := redis.Float64(client.Do("ZSCORE", key, userId))
	// if err != nil {
	// 	fmt.Println("ZSCORE2 error:", err)
	// 	return err
	// }

	fmt.Println("key:", key)
	fmt.Println("ov:", ov)
	fmt.Println("v:", v)

	if v == ov && ov != 0 {
		// 已经投过票了，不能重复投票
		fmt.Println("已经投过票了，不能重复投票")
		return errors.New("已经投过票了，不能重复投票")
	}

	var op float64

	if v > ov { // 改成赞成票
		op = 1
	} else { // 改成反对票
		op = -1
	}

	// 计算差值
	diffAbs := math.Abs(ov - v)

	// 打开事务
	client.Send("MULTI")
	// 更新分数（增加分数）
	client.Send("ZINCRBY", KeyPostScoreZSet, VoteScore*op*diffAbs, postId)
	// 记录用户为该帖子投票的数据
	if v == 0 {
		// 取消投票， 就要删除用户的投票记录
		client.Send("ZREM", key, userId)
		// 修改
		if ov == 432 { // 原本投赞成票
			client.Send("HINCRBY", KeyPostInfoHashPrefix+strconv.Itoa(int(postId)), "likes", -1)
		} else if ov == -432 { // 原本投反对票
			client.Send("HINCRBY", KeyPostInfoHashPrefix+strconv.Itoa(int(postId)), "unlikes", -1)
		}
	} else {
		// 更新投票纪录
		client.Send("ZADD", key, v, userId)
		if *direction == 1 {
			client.Send("HINCRBY", KeyPostInfoHashPrefix+strconv.Itoa(int(postId)), "likes", 1)
			if ov == -432 { // 原本投的是反对票
				client.Send("HINCRBY", KeyPostInfoHashPrefix+strconv.Itoa(int(postId)), "unlikes", -1)
			}
		} else {
			client.Send("HINCRBY", KeyPostInfoHashPrefix+strconv.Itoa(int(postId)), "unlikes", 1)
			if ov == 432 { // 原本投的是赞成票
				client.Send("HINCRBY", KeyPostInfoHashPrefix+strconv.Itoa(int(postId)), "likes", -1)
			}
		}
	}

	rc, err := client.Do("EXEC")
	fmt.Println("Redis voteForPost Transaction's result ", rc)

	return err
}

// 根据ids查询每篇帖子的投赞成票的数据
func GetPostVoteData(ids []string) (data []int64, err error) {
	client := RedisPool.Get()
	defer client.Close()

	// 遍历所有id，统计投赞成票的个数()
	for _, id := range ids {
		key := KeyCommunityPostSetPrefix + id
		client.Send("ZCOUNT", key, "1", "1")
	}

	cm, err := client.Do("Exec")
	// 将返回的结果转换为[]interface{}
	cmders, err := redis.Values(cm, err)
	if err != nil {
		return nil, err
	}

	data = make([]int64, 0, len(cmders))
	for _, cmder := range cmders {
		v, _ := redis.Int64(cmder, err)
		data = append(data, v)
	}
	return
}
