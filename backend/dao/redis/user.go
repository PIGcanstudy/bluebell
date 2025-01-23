package redis

import "fmt"

func StoreJWTToken(userId uint64, atoken string, rtoken string) error {
	client := RedisPool.Get()
	defer client.Close()

	key := fmt.Sprintf(keyUserTokenHashPrefix+"%d", userId)
	_, err := client.Do("HMSET", key, "atoken", atoken, "rtoken", rtoken)
	if err != nil {
		return err
	}

	return nil
}

func DeleteJWTToken(userId uint64) error {
	client := RedisPool.Get()
	defer client.Close()

	key := fmt.Sprintf(keyUserTokenHashPrefix+"%d", userId)
	_, err := client.Do("HDEL", key, "atoken", "rtoken")
	if err != nil {
		return err
	}

	return nil
}

func UpdateToken(userId uint64, atoken string, rtoken string) error {
	client := RedisPool.Get()
	defer client.Close()
	key := fmt.Sprintf(keyUserTokenHashPrefix+"%d", userId)
	_, err := client.Do("HMSET", key, "atoken", atoken, "rtoken", rtoken)
	if err != nil {
		fmt.Println("redis Usergo's UpdateToken func error: ", err)
		return err
	}

	return nil
}
