package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"smart-ledger-server/internal/config"
	"smart-ledger-server/internal/pkg/response"
	"smart-ledger-server/pkg/errcode"
)

// UserExistsChecker 用户存在性检查接口
type UserExistsChecker interface {
	ExistsByID(ctx context.Context, id uint64) (bool, error)
}

// Auth JWT认证中间件
func Auth(cfg *config.JWTConfig, userChecker UserExistsChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errcode.ErrTokenInvalid
			}
			return []byte(cfg.Secret), nil
		})

		if err != nil {
			if err == jwt.ErrTokenExpired {
				response.Error(c, errcode.ErrTokenExpired)
			} else {
				response.Error(c, errcode.ErrTokenInvalid)
			}
			c.Abort()
			return
		}

		// 验证Token
		if !token.Valid {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 获取Claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		// 获取用户ID
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		userID := uint64(userIDFloat)

		// 验证用户是否存在
		exists, err := userChecker.ExistsByID(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, errcode.ErrServer)
			c.Abort()
			return
		}
		if !exists {
			response.Error(c, errcode.ErrUserNotFound)
			c.Abort()
			return
		}

		// 设置用户ID到上下文
		c.Set("user_id", userID)
		c.Next()
	}
}
