package app

import (
	"errors"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"time"
)

var JwtSecret string

func ParseToken(tokenString string) (uint, uint, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JwtSecret), nil
	})

	if err != nil {
		return 0, 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if int64(claims["exp"].(float64)) < time.Now().Unix() {
			return 0, 0, "", errors.New("Token expired")
		}

		var userID uint = 0
		var adminID uint = 0
		var role string = ""

		if claimUserID, ok := claims["userID"]; ok {
			userID = uint(claimUserID.(float64))
		}

		if claimAdminID, ok := claims["adminID"]; ok {
			adminID = uint(claimAdminID.(float64))
		}

		if claimRole, ok := claims["role"]; ok {
			role = claimRole.(string)
		}

		return userID, adminID, role, nil
	}
	return 0, 0, "", err
}
