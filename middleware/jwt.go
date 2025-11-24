package middleware

import (
	"fmt"
	"time"
	"xzdp/config"
	"xzdp/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type userClaims struct {
	UserId               int64
	Phone                string
	jwt.RegisteredClaims // v5版本新加的方法
}

// 避免在 JWT 的 payload 中存储敏感的用户信息。因为 JWT 通常是可解码的，虽然签名可以保证其完整性，
// 但不能保证其保密性。如果需要存储一些用户相关的信息，可以使用加密的方式存储在服务器端，并在 JWT 中存储一个引用或标识符。
// 所以要对号码进行加密，或者使用其他不敏感的信息。

// 生成Token
func GenerateToken(phone string, userId int64) (string, error) {
	// 1. 接收手机号和userId
	// 2. 创建userClaims对象
	claims := userClaims{
		Phone:  phone,  //手机号
		UserId: userId, //用户ID
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.JwtOption.Expire)), //过期时间
			Issuer:    config.JwtOption.Issuer,                                     //签发者
			NotBefore: jwt.NewNumericDate(time.Now()),                              //生效时间
		},
	}
	// 3. 使用 HS256 算法和配置中的密钥对 claims 进行签名
	tokenStruct := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 4. 获得完整的签名的令牌
	return tokenStruct.SignedString([]byte(config.JwtOption.Secret))
}

// 解析Token
func ParseToken(token string) (*userClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &userClaims{}, func(token *jwt.Token) (any, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.JwtOption.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*userClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}

const (
	CtxKeyUserPhone       = "userPhone"
	CtxKeyUserId          = "userId"
	CtxKeyIsAuthenticated = "isAuthenticated"
)

// OptionalJWT 作为可选验证，无论是否提供令牌都会放行请求，但会在上下文中标记认证状态
func OptionalJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头的Authorization获取令牌
		token := c.GetHeader("Authorization")
		// 2. 没token就直接放行，有token就判断，反正这是可选校验
		if token == "" {
			// 未提供token，设置未认证状态
			c.Set(CtxKeyIsAuthenticated, false)
			// 放行请求，c.Abort()是拦截
			c.Next()
			return
		}

		// 2. 处理Bearer Token格式（去掉前缀"Bearer"）
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		} else {
			// token格式错误，设置未认证状态
			c.Set(CtxKeyIsAuthenticated, false)
			c.Next()
			return
		}

		claims, err := ParseToken(token)
		if err != nil {
			// token解析失败，设置未认证状态
			c.Set(CtxKeyIsAuthenticated, false)
			c.Next()
			return
		}

		// 验证关键字段是否为空
		if claims.Phone == "" || claims.UserId == 0 {
			// 字段为空，设置未认证状态
			c.Set(CtxKeyIsAuthenticated, false)
			c.Next()
			return
		}

		// token有效且字段完整，设置用户信息
		c.Set(CtxKeyUserPhone, claims.Phone)
		c.Set(CtxKeyUserId, claims.UserId)
		c.Set(CtxKeyIsAuthenticated, true)
		c.Next()
	}
}

// RequireAuth 作为强制验证，检查上下文中的认证状态，未认证则返回错误并拦截请求
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAuthenticated := c.GetBool(CtxKeyIsAuthenticated)
		if !isAuthenticated {
			response.Error(c, response.ErrInvalidAuthHeader, "请先登录")
			c.Abort()
			return
		}
		c.Next()
	}
}
