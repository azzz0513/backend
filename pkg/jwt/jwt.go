package jwt

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"time"
)

var mySecret = []byte("secret")

// MyClaims 自定义声明结构体并内嵌jwt，StandardClaims
// jwt包自带的jwt.StandardClaims只定义了官方字段
// 我们这里需要额外记录一个username字段，所以要自定义结构体
// 如果想要保存更多信息，都可以添加到这个结构体中
type MyClaims struct {
	UserID int64 `json:"user_id"`
	jwt.StandardClaims
}

type CheckinClaims struct {
	CheckinID int64 `json:"checkin_id"`
	jwt.StandardClaims
}

// GenToken 生成access token和refresh token
func GenToken(userID int64) (string, error) {
	// 创建一个我们自己的声明
	c := MyClaims{
		UserID: userID, // 自定义字段
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(viper.GetDuration("auth.jwt_expire") * time.Hour).Unix(), // 过期时间
			Issuer:    "shit",                                                                  // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	// 使用指定的secret签名并获得完整的编码后的字符串token
	return token.SignedString(mySecret)
}

// GenCheckinToken 生成打卡活动token
func GenCheckinToken(checkinID int64, duration uint) (string, error) {
	// 创建声明
	c := CheckinClaims{
		CheckinID: checkinID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(duration) * time.Minute).Unix(), // 过期时间
			Issuer:    "shit",                                                       // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(mySecret)
}

// ParseToken 解析access token
func ParseToken(tokenString string) (claims *MyClaims, err error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func ParseCheckinToken(tokenString string) (claims *CheckinClaims, err error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &CheckinClaims{}, func(token *jwt.Token) (interface{}, error) {
		return mySecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CheckinClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
