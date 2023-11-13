package app

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
    "github.com/stripe/stripe-go/v72/webhook"
	"github.com/jinzhu/gorm"
	"fmt"
	"time"
	"log"
	"io"
	"encoding/json"
)

type CheckoutSession struct {
	ID                 string `json:"id"`
	Object             string `json:"object"`
	AmountSubtotal     int    `json:"amount_subtotal"`
	AmountTotal        int    `json:"amount_total"`
	Currency           string `json:"currency"`
	Customer           string `json:"customer"`
	CustomerDetails    struct {
		Address struct {
			City    string `json:"city"`
			Country string `json:"country"`
			Line1   string `json:"line1"`
			Line2   string `json:"line2"`
		} `json:"address"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"customer_details"`
	PaymentIntent string `json:"payment_intent"`
	Livemode      bool   `json:"livemode"`
	Mode          string `json:"mode"`
}


func PaymentStripe(c *gin.Context) {
	stripe.Key = StripeKeyTest

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String("price_1O2tQkEe01UiubRXXxGjEr73"),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(SuccessURL),
		CancelURL:  stripe.String(FailURL),
	}

	sess, err := session.New(params)
	if err != nil {
		log.Printf("session.New: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

    now := time.Now()
	date := now.Format("2006-01-02")

	db := c.MustGet("db").(*gorm.DB)
	userID, _ := GetUserC(c)

	user, err := GetUser(c, db, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка при получении пользователя: %v", err)})
		return
	}

	orderID, err := GetNewOrder(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка при получении последнего заказа: %v", err)})
		return
	}

	transaction := Transaction{
		UserID:          userID,
		Email:           user.Email,
		Type:            "Stripe",
		TransactionType: "Pay",
		OrderID:         orderID,
		PaymentId:       sess.ID,
		Status:          "pending",
		Date:            date,
		Taken:           false,
	}

	result := db.Create(&transaction)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Ошибка при сохранении подписки: %v", result.Error)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessionURL": sess.URL})
}

func StripeWebhook(c *gin.Context) {

    log.Println("[StripeWebhook] Started")

	signatureHeader := c.GetHeader("Stripe-Signature")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("[StripeWebhook] Error reading body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	event, err := webhook.ConstructEvent(body, signatureHeader, PaymentStripeWebhookKey)
	if err != nil {
		log.Println("[StripeWebhook] Failed to verify webhook signature")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to verify webhook signature"})
		return
	}
    // log.Printf("Full event details: %s\n", string(event.Data.Raw))

	if event.Type == "checkout.session.completed" {
		checkoutSession := &CheckoutSession{}
		err := json.Unmarshal(event.Data.Raw, checkoutSession)
		if err != nil {
			log.Println("[StripeWebhook] Error parsing webhook JSON")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse webhook JSON"})
			return
		}

        log.Printf("[StripeWebhook] Parsed checkout session data: %+v", checkoutSession)

        totalAmount := checkoutSession.AmountTotal
        if totalAmount != 1800 {
            log.Printf("[StripeWebhook] Unexpected total amount: %d", totalAmount)
            c.JSON(http.StatusBadRequest, gin.H{"error": "Unexpected total amount"})
            return
        }

        db := c.MustGet("db").(*gorm.DB)

        tx := db.Begin()
        if tx.Error != nil {
            log.Println("[StripeWebhook] Error starting transaction")
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }

        var transaction Transaction
        if err := tx.Where("payment_id = ?", checkoutSession.ID).First(&transaction).Error; err != nil {
            tx.Rollback()
            log.Println("[StripeWebhook] Error finding transaction")
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }

        transaction.Status = "success"
        transaction.Taken = true
        if err := tx.Save(&transaction).Error; err != nil {
            tx.Rollback()
            log.Printf("[StripeWebhook] Error updating transaction: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }

        var user User
        if err := tx.Where("id = ?", transaction.UserID).First(&user).Error; err != nil {
            tx.Rollback()
            log.Printf("[StripeWebhook] Error finding user: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }

        now := time.Now()
        userAccessTo := user.AccessTo
        if userAccessTo.Before(now) {
            user.AccessTo = now.AddDate(0, 0, AccessDays)
        } else {
            user.AccessTo = userAccessTo.AddDate(0, 0, AccessDays)
        }
        if err := tx.Save(&user).Error; err != nil {
            tx.Rollback()
            log.Printf("[StripeWebhook] Unhandled event type: %s", event.Type)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }

        if err := tx.Commit().Error; err != nil {
            log.Printf("Error committing transaction: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"status": "success"})
    } else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unhandled event type"})
	}
}

func SubscribeStripe(c *gin.Context) {

    stripe.Key = StripeKeyTest

    params := &stripe.CheckoutSessionParams{
        Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
        PaymentMethodTypes: stripe.StringSlice([]string{
            "card",
        }),
        LineItems: []*stripe.CheckoutSessionLineItemParams{
            {
                Price:    stripe.String("price_1NuBbuEe01UiubRXutkK0TzS"),
                Quantity: stripe.Int64(1),
            },
        },
        SuccessURL: stripe.String(SuccessURL),
        CancelURL:  stripe.String(FailURL),
    }

    sess, err := session.New(params)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"sessionId": sess.ID})
}
