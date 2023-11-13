package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"time"
)

func Subscribe(c *gin.Context) {

	var request struct {
		PaymentMethod string `json:"payment_method"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paymentMethod := request.PaymentMethod

	var data gin.H
	var err error

	switch paymentMethod {
	case "MIR":
		data, err = MIRSubscribe(c, paymentMethod)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "Stripe":

	case "PayPal":

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method"})
		return
	}

	c.JSON(http.StatusOK, data)
}

func SubscribeCrypto(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID, _ := GetUserC(c)
	user, err := GetUser(c, db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("User not found: %v", err)})
		return
	}
	var request struct {
		PaymentMethod   string `json:"payment_method"`
		TransactionHash string `json:"transaction_hash"`
		UserAddress     string `json:"user_address"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}
	if user.Subscribe != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Your subscribe"})
		return
	}
	user.Address = request.UserAddress
	result := db.Save(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save user: %v", result.Error)})
		return
	}

	var data gin.H
	switch request.PaymentMethod {
	case "USDT polygon", "USDC polygon":
		data, err = CryptoSubscribePolygon(c, request.TransactionHash, request.PaymentMethod)
		fmt.Println(" data ", data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "USDT binance", "USDC binance":

	case "USDT tron", "USDC tron":

	case "USDT solana", "USDC solana":

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func UnSubscribe(c *gin.Context) {

	db := c.MustGet("db").(*gorm.DB)
	userID, _ := GetUserC(c)
	user, err := GetUser(c, db, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}
	user.Subscribe = ""
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("unable to update user: %v", err).Error()})
		return
	}

	var message string

	switch user.Subscribe {
	case "MIR":
		message, err = MIRUnSubscribe(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	case "Stripe":

	case "PayPal":

	default:
		c.JSON(http.StatusOK, gin.H{"error": "UnSubscribe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func AutoRenewSubscriptions(db *gorm.DB) {

	for {
		var users []User
		db.Where("subscribe <> ?", "").Find(&users)

		for _, user := range users {
			currentTime := time.Now()

			renewalTime := currentTime.Add(-3 * time.Hour)
			if user.AccessTo.Before(renewalTime) {
				switch user.Subscribe {
				case "USDT polygon", "USDC polygon":
					AutoPaymentEVM(PolygonID, PolygonContract, PolygonPrivateKey, PolygonProvider, &user, db)
				case "MIR":

				}

			}
		}
		time.Sleep(2 * time.Hour)
	}
}
