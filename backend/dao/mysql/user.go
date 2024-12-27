package mysql

import (
	"bluebell/models"
	"crypto/md5"
	"encoding/hex"
	"fmt"
)

func EncryptPassword(Password string) string {
	m := md5.New()
	m.Write([]byte(Password))
	re := m.Sum(nil)
	// 将md5加密后的结果转换为16进制字符串
	return hex.EncodeToString(re)
}

func CheckUserExist(username string) bool {
	var user models.User
	DB.Table("user").Where("username =?", username).First(&user)

	return user.Username == username
}

// 返回-1表示数据库查询出错, -2表示用户不存在，返回0表示查询结果不一致。返回1表示一致
func CheckPassword(data models.LoginForm) int {
	var password string
	re := DB.Table("user").Select("password").Where("username =?", data.Username).Find(&password)
	if re.Error != nil {
		fmt.Println(re.Error)
		return -1
	}
	var pwd string = EncryptPassword(data.Password)

	if password != pwd && password != "" {
		return 0
	}

	if password == "" {
		return -2
	}

	return 1
}
