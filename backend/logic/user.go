package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/jwt"
	"bluebell/pkg/snowflake"
	"fmt"
)

func Signup(data models.RegisterForm) bool {
	// 首先查看用户是否已存在
	isExist := mysql.CheckUserExist(data.Username)

	fmt.Println(isExist)

	if isExist {
		return false
	}
	// 不存在就创建用户进数据库

	// 先使用雪花算法生成UUID
	userId := snowflake.GetID()

	var user models.User
	user.User_id = uint64(userId)
	user.Username = data.Username
	user.Password = data.Password

	if err := InsertUser(user); err != nil {
		fmt.Println("用户插入数据库失败， 原因为：", err)
		return false
	}
	return true
}

func InsertUser(user models.User) error {

	user.Password = mysql.EncryptPassword(user.Password)
	fmt.Println(user)
	// 插入数据库中
	result := mysql.DB.Create(&user)

	if result.Error != nil {
		return result.Error
	}
	return nil
}

func Signin(data models.LoginForm) (user models.User, isSuccess int) {
	// 查询数据库中的用户数据
	isSuccess = mysql.CheckPassword(data, &user)
	if isSuccess != 1 {
		if isSuccess == 0 {
			fmt.Println("用户名或密码错误")
		} else if isSuccess == -1 {
			fmt.Println("查询数据库出错")
		} else {
			fmt.Println("用户不存在")
		}
		return
	}

	// 生成JWTTOKEN

	atoken, rotken, err := jwt.GenToken(uint64(user.User_id), user.Username)

	if err != nil {
		fmt.Println("生成JWTTOKEN失败， 原因为：", err)
		return
	}

	user.AccessToken = atoken
	user.RefreshToken = rotken

	// 保存JWTToken存到redis中（这样是否违背了jwt的设计理念）
	if err := redis.StoreJWTToken(user.User_id, atoken, rotken); err != nil {
		fmt.Println("保存JWTToken到redis失败， 原因为：", err)
		return
	}

	// 等于1表示密码与用户匹配
	fmt.Println("登录成功")

	return
}

func SignOut(userId uint64) error {
	// 从redis中删除JWTToken
	if err := redis.DeleteJWTToken(userId); err != nil {
		fmt.Println("从redis中删除JWTToken失败， 原因为：", err)
		return err
	}

	fmt.Println("登出成功")
	return nil
}

func UpdateToken(userId uint64, accessToken string, refreshToken string) error {
	// 更新redis中的JWTToken
	if err := redis.StoreJWTToken(userId, accessToken, refreshToken); err != nil {
		fmt.Println("从redis中更新JWTToken失败， 原因为：", err)
		return err
	}

	fmt.Println("更新JWTToken成功")
	return nil
}
