package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
)


var SubscribePriceCrypto int
var SubscribeDescription string
var PaymentDescription string
var RenewalDescription string
var RenewalInitDescription string
var SuccessURL string
var FailURL string
var TestDays int
var AccessDays int

var PolygonID int
var PolygonDecimal int
var PayPriceCrypto int
var PolygonUSDT string
var PolygonUSDC string
var PolygonContract string

type Transaction struct {
	gorm.Model
	UserID          uint
	Email           string
	Type            string
	TransactionType string
	PaymentId       string
	OrderID         uint
	Amount          int
	Status          string
	PaymentURL      string
	Taken           bool
	Date            string
}

func PaymentMIR(c *gin.Context) {

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
		data, err = MIRPayment(c, paymentMethod)
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

func PaymentCrypto(c *gin.Context) {

	// Парсинг входящих данных
	var request struct {
		PaymentMethod   string `json:"payment_method"`
		TransactionHash string `json:"transactionHash"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	var data gin.H
	var err error

	switch request.PaymentMethod {
	case "USDT polygon", "USDC polygon":
		data, err = CryptoPaymentPolygon(c, request.TransactionHash, request.PaymentMethod)
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

func UserTransactionsUpdated(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID, _ := GetUserC(c)
	user, err := GetUser(c, db, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	var transactions []Transaction
	if err := db.Where("user_id = ? AND taken = ?", userID, false).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find transactions"})
		return
	}

	for _, transaction := range transactions {
		switch transaction.Type {
		case "USDT polygon", "USDC polygon":
			err := UpdateTransactionEVM(&transaction, db, &user)
			if err != nil {
				fmt.Println("error: Failed to update EVM transaction")
			}
		case "MIR":
			err := UpdateTransactionMIR(&transaction, db, &user)
			if err != nil {
				fmt.Println("error: Failed to update MIR transaction")
			}
		case "Stripe":

		case "PayPal":

		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "Transactions updated"})
}
