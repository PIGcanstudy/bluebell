package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gopkg.in/ini.v1"
)

// MyClaims 自定义声明结构体并内嵌jwt.StandardClaims
// jwt包自带的jwt.StandardClaims只包含了官方字段
// 我们这里需要额外记录一个UserID字段，所以要自定义结构体
// 如果想要保存更多信息，都可以添加到这个结构体中
type MyClaims struct {
	UserId   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

var mySecret = []byte("PigCanStudy")

// 返回用于生成的密码
func keyFunc(*jwt.Token) (i interface{}, e error) {
	return mySecret, nil
}

// 定义JWT的过期时间
const TokenExpireDuration = time.Hour * 2

// 生成AccessToken和RefreshToken
func GenToken(userId uint64, username string) (aToken, rToken string, err error) {
	cfg, _ := ini.Load("./conf/config.ini")
	duration, _ := cfg.Section("auth").Key("jwt_expire").Int()
	// 创建一个自己的声明
	c := MyClaims{
		userId,     // 自定义字段
		"username", // 自定义字段
		jwt.StandardClaims{ // JWT规定的7个官方字段
			ExpiresAt: time.Now().Add(
				time.Duration(duration) * time.Hour).Unix(), // 过期时间
			Issuer: "bluebell", // 签发人
		},
	}
	// 加密并获得完整的编码后的字符串token
	aToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret)

	// refresh token 不需要存任何自定义数据
	rToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7).Unix(), // 过期时间(设置为7天)
		Issuer:    "bluebell",                                // 签发人
	}).SignedString(mySecret)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return
}

// GenToken 生成 Token
func GenToken2(userID uint64, username string) (Token string, err error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		userID,     // 自定义字段
		"username", // 自定义字段
		jwt.StandardClaims{ // JWT规定的7个官方字段
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(), // 过期时间
			Issuer:    "bluebell",                                 // 签发人
		},
	}
	// 加密并获得完整的编码后的字符串token
	Token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(mySecret)

	// refresh token 不需要存任何自定义数据
	//rToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
	//	ExpiresAt: time.Now().Add(time.Second * 30).Unix(), // 过期时间
	//	Issuer:    "bluebell",                              // 签发人
	//}).SignedString(mySecret)	// 使用指定的secret签名并获得完整的编码后的字符串token
	return
}

func ParseToken(tokenString string) (claims *MyClaims, err error) {
	// 解析token
	var token *jwt.Token
	claims = new(MyClaims)
	token, err = jwt.ParseWithClaims(tokenString, claims, keyFunc)
	fmt.Printf("Parsed ExpiresAt: %v\n", time.Unix(claims.ExpiresAt, 0))
	if err != nil {
		return
	}
	if !token.Valid { // 校验token是否有效

		err = errors.New("invalid token")
	}
	return
}

// RefreshToken 刷新AccessToken
func RefreshToken(aToken, rToken string) (newAToken, newRToken string, err error) {
	// refresh token无效直接返回
	if _, err = jwt.Parse(rToken, keyFunc); err != nil {
		return
	}

	// 从旧access token中解析出claims数据	解析出payload负载信息
	var claims MyClaims
	_, err = jwt.ParseWithClaims(aToken, &claims, keyFunc)
	v, _ := err.(*jwt.ValidationError)

	// 当access token是过期错误 并且 refresh token没有过期时就创建一个新的access token
	if v.Errors == jwt.ValidationErrorExpired {
		return GenToken(claims.UserId, claims.Username)
	}
	return
}
