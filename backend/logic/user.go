package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/snowflake"
	"fmt"
)

func Signup(data models.RegisterForm) bool {
	// 首先查看用户是否已存在
	isExist := mysql.CheckUserExist(data.Username)

	if isExist {
		return false
	}
	// 不存在就创建用户进数据库

	// 先使用雪花算法生成UUID
	userId := snowflake.GetID()

	var user models.User
	user.User_id = userId
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
	fmt.Println(result)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func Signin(data models.LoginForm) int {
	// 查询数据库中的用户数据
	isSuccess := mysql.CheckPassword(data)
	if isSuccess != 1 {
		if isSuccess == 0 {
			fmt.Println("用户名或密码错误")
		} else if isSuccess == -1 {
			fmt.Println("查询数据库出错")
		} else {
			fmt.Println("用户不存在")
		}
		return isSuccess
	}

	// 等于1表示密码与用户匹配
	fmt.Println("登录成功")

	return isSuccess
}