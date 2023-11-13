package app

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"strings"
	"time"
)

type User struct {
	gorm.Model
	Phone             string
	Email             string `gorm:"unique"`
	Address           string
	Password          string
	AccessTo          time.Time
	Subscribe         string
	PaymentId         string
	Confirmed         bool
	ConfirmedCode     string
	ConfirmationTries int
	Block             bool
}

func GetUserC(c *gin.Context) (uint, error) {

	userID, exists := c.Get("userID")
	if !exists {
		return 0, errors.New("User id not found")
	}

	id, ok := userID.(uint)
	if !ok {
		return 0, errors.New("User id has wrong type")
	}

	return id, nil
}

func CheckToken(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header not provided"})
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header format must be Bearer <token>"})
		return
	}
	_, _, _, err := ParseToken(parts[1])
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "Invalid authorization token"})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func GetUser(c *gin.Context, db *gorm.DB, userID uint) (User, error) {
	user := User{}
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return User{}, err
	}
	return user, nil
}

func GetNewOrder(db *gorm.DB) (uint, error) {
	var maxOrderID *uint

	row := db.Table("transactions").Select("MAX(order_id)").Row()
	err := row.Scan(&maxOrderID)

	if err != nil {
		return 250, errors.New("Ошибка при получении максимального значения OrderID")
	}

	if maxOrderID == nil {

		fmt.Println("Нет строк в таблице или значений order_id")
		return 250, nil
	}

	newOrderID := *maxOrderID + 1

	return newOrderID, nil
}

func CheckAccess(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID, _ := GetUserC(c)
	user, err := GetUser(c, db, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	now := time.Now()

	var success bool
	if user.AccessTo.IsZero() {
		success = false
	} else {
		success = now.Before(user.AccessTo)
	}

	response := gin.H{
		"access_to": user.AccessTo.Format("2006-01-02"),
		"subscribe": user.Subscribe,
		"success":   success,
	}
	c.JSON(http.StatusOK, response)
}
