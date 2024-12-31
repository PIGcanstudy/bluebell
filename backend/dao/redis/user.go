package redis

func StoreToken(token string, userId int64) {
	conn := RedisPool.Get()
	defer conn.Close()
	conn.Do("SET", userId, token)

}
