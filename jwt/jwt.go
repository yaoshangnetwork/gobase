package jwt

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	jwtgo "github.com/golang-jwt/jwt/v5"
)

// 中间件
func NewGinMiddleware(jwtSecret string, keys ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			ctx.Abort()
			return
		}

		token, err := jwtgo.Parse(tokenString, func(token *jwtgo.Token) (any, error) {
			if _, ok := token.Method.(*jwtgo.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}

		if claims, ok := token.Claims.(jwtgo.MapClaims); ok {
			// 检查过期时间
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
					ctx.Abort()
					return
				}
			}

			// 存值
			for _, key := range keys {
				if v, ok := claims[key]; ok {
					ctx.Set(key, v)
				}
			}
		}

		ctx.Next()
	}
}

// 生成 JWT
func GenerateToken(jwtSecret string, exp time.Time, data map[string]any) (string, error) {
	claims := jwtgo.MapClaims{
		"exp": exp.Unix(),        // 过期时间
		"iat": time.Now().Unix(), // 签发时间
	}
	for k, v := range data {
		if _, ok := claims[k]; !ok {
			claims[k] = v
		}
	}
	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
