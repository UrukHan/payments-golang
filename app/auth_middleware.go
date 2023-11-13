package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header not provided"})
			return
		}

		// Split the authHeader string into two parts.
		// Normally Authorization header should be of format `Bearer <token>`.
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer <token>"})
			return
		}

		// parts[1] should contain the token only.
		userID, adminID, role, err := ParseToken(parts[1])

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			return
		}

		// сохраняем ID пользователя и роль в контексте, чтобы их можно было использовать в последующих обработчиках
		c.Set("userID", userID)
		c.Set("adminID", adminID)
		c.Set("role", role)

		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, присутствует ли "role" в контексте
		role, exists := c.Get("role")
		if !exists || role.(string) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access restricted to administrators only"})
			return
		}

		c.Next()
	}
}
