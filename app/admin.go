package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"time"
)

var AdminPassword string
var AdminEmail string
var TerminalKey string
var TerminalPass string
var PolygonProvider string
var PolygonPrivateKey string
var PolygonOwner string
var StripeKey string
var StripeKeyTest string
var PaymentStripeWebhookKey string

func GetUserData(c *gin.Context) {

	var requestData struct {
		User string `json:"user"`
	}
	fmt.Println("START")
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request data"})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	var user User
	if err := db.Where("email = ?", requestData.User).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	response := gin.H{
		"access":    user.AccessTo,
		"subscribe": user.Subscribe,
		"block":     user.Block,
	}

	c.JSON(http.StatusOK, response)
}

func ChangeUserData(c *gin.Context) {

	var requestData struct {
		Address   string `json:"address"`
		AccessTo  string `json:"accessTo"`
		Subscribe string `json:"subscribe"`
		Block     bool   `json:"block"`
	}

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request data"})
		fmt.Println(err)
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	var user User
	if err := db.Where("email = ?", requestData.Address).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	timeFormat := "2006-01-02"
	accessTo, err := time.Parse(timeFormat, requestData.AccessTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	user.AccessTo = accessTo
	user.Subscribe = requestData.Subscribe
	user.Block = requestData.Block

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "User data updated successfully"})
}

func GetData(c *gin.Context) {

	var requestData struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request data"})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	var transactions []Transaction

	if err := db.Where("created_at BETWEEN ? AND ?", requestData.StartDate, requestData.EndDate).Order("created_at desc").Find(&transactions).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "No payment information found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to get payment information"})
		return
	}

	response := make([]gin.H, len(transactions))
	for i, transaction := range transactions {
		response[i] = gin.H{
			"paymentId": transaction.PaymentId,
			"email":     transaction.Email,
			"status":    transaction.Status,
			"date":      transaction.CreatedAt.Format("2006-01-02 15:04"),
		}
	}
	c.JSON(http.StatusOK, response)
}

func CancelPayment(c *gin.Context) {
	var requestData struct {
		PaymentId string `json:"PaymentId"`
	}
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request data"})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	var transaction Transaction
	if err := db.Where("payment_id = ?", requestData.PaymentId).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch payment from database"})
		return
	}

	responseCancel, err := requestCancel(requestData.PaymentId)

	if err != nil {
		fmt.Println("Ошибка при отмене платежа:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при отмене платежа"})
		return
	}

	if responseCancel.Success {

		transaction.Status = responseCancel.Status
		if err := db.Save(&transaction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to update payment status in database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":   responseCancel.Success,
			"status":    responseCancel.Status,
			"paymentId": responseCancel.PaymentId,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to cancel payment"})
	}
}

func RemovePayment(c *gin.Context) {
	var requestData struct {
		PaymentId string `json:"PaymentId"`
	}
	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request data"})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	var transaction Transaction
	if err := db.Where("payment_id = ?", requestData.PaymentId).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	if err := db.Delete(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Transaction removed successfully"})
}
