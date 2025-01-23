package models

import (
	"encoding/json"
	"errors"
)

type SimpleUser struct {
	User_Id  int64  `json:"user_id"`
	Username string `json:"username"`
}

type User struct {
	Id           int    `json:"id"`
	User_id      uint64 `json:"user_id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Email        string `json:"email"`
	AccessToken  string `json:"access_token" gorm:"-"`
	RefreshToken string `json:"refresh_token" gorm:"-"`
}

func (User) TableName() string {
	return "user"
}

// 由于前端要传数据并且是Json 数据所以需要反序列化
// 前端发送登录请求的时候需要用到
func (u *User) Unmarshal(data []byte) (err error) {
	required := struct {
		Username string `json:"username" db:"username"`
		Password string `json:"password" db:"password"`
	}{}

	// 反序列化
	if err = json.Unmarshal(data, &required); err != nil {
		return
	}

	// 错误处理
	if required.Username == "" {
		err = errors.New("username is empty")
		return
	}

	if required.Password == "" {
		err = errors.New("password is empty")
		return
	}

	// 赋值给结构体
	u.Username = required.Username
	u.Password = required.Password

	return

}

type RegisterForm struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binging:"required"`
	Confirm  string `json:"confirm" binding:"required" equalfield:"Password"`
}

// 反序列化 注册用
func (rf *RegisterForm) UnmarshalJson(data []byte) (err error) {
	required := struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Confirm  string `json:"confirm" binding:"required, eqfield=Password"`
	}{}

	//  反序列化
	if err = json.Unmarshal(data, &required); err != nil {
		return
	}

	// 错误处理
	if required.Username == "" {
		err = errors.New("username is empty")
		return
	}

	if required.Password == "" {
		err = errors.New("password is empty")
		return
	}

	if required.Confirm != required.Password {
		err = errors.New("confirm password is not equal to password")
		return
	}

	// 赋值给结构体
	rf.Username = required.Username
	rf.Password = required.Password
	rf.Confirm = required.Confirm

	return
}

type LoginForm struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (lf *LoginForm) UnmarshalJson(data []byte) (err error) {
	var required LoginForm

	// 反序列化
	if err = json.Unmarshal(data, &required); err != nil {
		return
	}

	lf.Password = required.Password
	lf.Username = required.Username

	return
}
