package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
)

var PayPriceMIR int
var SubscribePriceMIR int

func MIRPayment(c *gin.Context, paymentMethod string) (gin.H, error) {

	now := time.Now()
	date := now.Format("2006-01-02")
	priceStr := strconv.Itoa(PayPriceMIR)

	db := c.MustGet("db").(*gorm.DB)
	userID, _ := GetUserC(c)
	user, err := GetUser(c, db, userID)
	if err != nil {
		return nil, fmt.Errorf("User not found: %v", err)
	}

	orderID, err := GetNewOrder(db)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении последнего заказа: %v", err)
	}

	initData := map[string]string{
		"Phone": user.Phone,
		"Email": user.Email,
	}

	receipt := map[string]interface{}{
		"Email":    user.Email,
		"Phone":    user.Phone,
		"Taxation": "patent",
		"Items": []map[string]interface{}{
			{
				"Name":     PaymentDescription,
				"Price":    priceStr,
				"Quantity": 1.00,
				"Amount":   priceStr,
				"Tax":      "vat0",
			},
		},
	}

	// Выполняем инициализацию платежа
	responseInit, err := requestInit(strconv.Itoa(int(orderID)), strconv.Itoa(int(userID)), priceStr, PaymentDescription, initData, receipt)
	if err != nil {
		fmt.Println("Ошибка при инициализации платежа:", err)
		return nil, fmt.Errorf("Ошибка при инициализации платежа: %v", err)
	}
	// Обновить статус оплаты
	if responseInit.Success == false {
		return nil, fmt.Errorf("Ошибка при инициализации платежа: %v", responseInit.Message)
	}

	// Сохранение данных о транзакции в БД (если у вас есть соответствующая таблица)
	transaction := Transaction{
		UserID:          userID,
		Type:            paymentMethod,
		TransactionType: "Pay",
		PaymentId:       responseInit.PaymentId,
		OrderID:         orderID,
		Amount:          PayPriceMIR,
		Status:          responseInit.Status,
		Taken:           false,
		Date:            date,
	}
	// Сохранить transaction в базе данных
	result := db.Create(&transaction)

	if result.Error != nil {
		return nil, fmt.Errorf("Ошибка при сохранении подписки: %v", result.Error)
	}

	data := gin.H{
		"url":           responseInit.PaymentURL,
		"paymentStatus": responseInit.Status,
		"date":          date,
	}
	return data, nil
}

func MIRSubscribe(c *gin.Context, paymentMethod string) (gin.H, error) {

	now := time.Now()
	date := now.Format("2006-01-02")
	priceStr := strconv.Itoa(SubscribePriceMIR)

	db := c.MustGet("db").(*gorm.DB)
	userID, _ := GetUserC(c)
	user, err := GetUser(c, db, userID)
	if err != nil {
		return nil, fmt.Errorf("User not found: %v", err)
	}

	orderID, err := GetNewOrder(db)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении последнего заказа: %v", err)
	}

	initData := map[string]string{
		"Phone": user.Phone,
		"Email": user.Email,
	}

	receipt := map[string]interface{}{
		"Email":    user.Email,
		"Phone":    user.Phone,
		"Taxation": "patent",
		"Items": []map[string]interface{}{
			{
				"Name":     PaymentDescription,
				"Price":    priceStr,
				"Quantity": 1.00,
				"Amount":   priceStr,
				"Tax":      "vat0",
			},
		},
	}

	responseInit, err := requestInit(strconv.Itoa(int(orderID)), strconv.Itoa(int(userID)), priceStr, PaymentDescription, initData, receipt)
	if err != nil {
		fmt.Println("Ошибка при инициализации платежа:", err)
		return nil, fmt.Errorf("Ошибка при инициализации платежа: %v", err)
	}

	if responseInit.Success == false {
		return nil, fmt.Errorf("Ошибка при инициализации платежа: %v", responseInit.Message)
	}

	transaction := Transaction{
		UserID:          userID,
		Type:            paymentMethod,
		TransactionType: "Subscribe",
		PaymentId:       responseInit.PaymentId,
		OrderID:         orderID,
		Amount:          SubscribePriceMIR,
		Status:          responseInit.Status,
		Taken:           false,
		Date:            date,
	}

	result := db.Create(&transaction)

	if result.Error != nil {
		return nil, fmt.Errorf("Ошибка при сохранении подписки: %v", result.Error)
	}

	data := gin.H{
		"url":           responseInit.PaymentURL,
		"paymentStatus": responseInit.Status,
		"date":          date,
	}

	return data, nil
}

func UpdateTransactionMIR(transaction *Transaction, db *gorm.DB, user *User) error {

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	responseState, err := requestState(transaction.PaymentId)
	if err != nil {
		tx.Rollback()
		return err
	}

	if responseState.Status == "CANCELED" || responseState.Status == "REFUNDED" || responseState.Status == "REVERSED" {
		transaction.Status = responseState.Status
		transaction.Taken = true
		if err := tx.Save(transaction).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else if transaction.Status == "CONFIRMED" {
		if transaction.TransactionType == "Pay" || transaction.TransactionType == "SubscribePay" {
			now := time.Now()

			var userAccessTo time.Time
			if !user.AccessTo.IsZero() {
				userAccessTo = user.AccessTo
			}

			if userAccessTo.Before(now) {
				user.AccessTo = now.AddDate(0, 0, 1)
			} else {
				user.AccessTo = userAccessTo.AddDate(0, 0, 1)
			}

			if err := tx.Save(user).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else if transaction.TransactionType == "Subscribe" {

		}
	}

	return tx.Commit().Error
}

func MIRUnSubscribe(c *gin.Context) (string, error) {

	return "Successfully unsubscribed", nil
}

//func TinkoffProcessSubscriptionInit(db *gorm.DB, subscribeInit Subscription, user User) (string, error) {
//	now := time.Now()
//	date := now.Format("2006-01-02")
//	initData := map[string]string{
//		"Phone": user.Phone,
//		"Email": user.Email,
//	}
//
//	adminData, err := ReadAdminDataFromFile()
//	if err != nil {
//		return "", fmt.Errorf("Ошибка чтения adminData")
//	}
//
//	receipt := map[string]interface{}{
//		"Email":    user.Email,
//		"Phone":    user.Phone,
//		"Taxation": "patent",
//		"Items": []map[string]interface{}{
//			{
//				"Name":     RenewalInitDescription,
//				"Price":    SubscribePriceRUB,
//				"Quantity": 1.00,
//				"Amount":   SubscribePriceRUB,
//				"Tax":      "vat0",
//			},
//		},
//	}
//
//	orderID, err := GetNewOrder(db)
//	if err != nil {
//		return "", fmt.Errorf("Ошибка при получении последнего заказа: %v", err)
//	}
//
//	// Выполняем инициализацию для инициализации платежей подписок
//	responseInit, err := requestSubscribeInit(adminData.TerminalKey, strconv.Itoa(int(orderID)),
//		subscribeInit.UserID, PayPriceRUB, RenewalInitDescription, initData, receipt)
//
//	if err != nil {
//		return "", fmt.Errorf("Ошибка при инициализации платежа: %v", err)
//	}
//
//	// Обновить статус оплаты
//	if responseInit.Success == false {
//		return "", fmt.Errorf("Ошибка при инициализации платежа")
//	}
//
//	if user.SubscribedData == "" {
//		now := time.Now()
//		user.SubscribedData = now.Add(1 * 24 * time.Hour).Format("2006-01-02")
//	}
//	user.Subscribe = "mirPay"
//
//	err = db.Transaction(func(tx *gorm.DB) error {
//		subscribeInit.Taken = true
//		if err := tx.Save(&subscribeInit).Error; err != nil {
//			return err
//		}
//		if err := tx.Save(&user).Error; err != nil {
//			return err
//		}
//
//		subscribe := Subscription{
//			UserID:     subscribeInit.UserID,
//			PaymentId:  responseInit.PaymentId,
//			Status:     responseInit.Status,
//			PaymentURL: responseInit.PaymentURL,
//			Taken:      false,
//			PreInit:    false,
//			Date:       date,
//		}
//		if err := tx.Create(&subscribe).Error; err != nil {
//			return err
//		}
//		return nil
//	})
//
//	if err != nil {
//		return "", fmt.Errorf("Ошибка при обновлении данных подписки и пользователя: %v", err)
//	}
//
//	return user.SubscribedData, nil
//}
//
//func TinkoffProcessSubscription(db *gorm.DB, subscribeInit Subscription, user User) (string, error) {
//	now := time.Now()
//	date := now.Format("2006-01-02")
//
//	adminData, err := ReadAdminDataFromFile()
//	if err != nil {
//		return "", fmt.Errorf("Ошибка чтения adminData")
//	}
//
//	err = db.Transaction(func(tx *gorm.DB) error {
//		if user.SubscribedData <= now.Format("2006-01-02") {
//			responseCardList, txErr := RequestGetCardList(adminData.TerminalKey, subscribeInit.UserID)
//			fmt.Println("responseCardList", responseCardList)
//			if txErr != nil {
//				return fmt.Errorf("Ошибка при запросе списка карт: %v", txErr)
//			}
//			responseCharge, txErr := RequestCharge(adminData.TerminalKey, subscribeInit.PaymentId, responseCardList.RebillId)
//			fmt.Println("responseCharge", responseCharge)
//			if txErr != nil {
//				return fmt.Errorf("Ошибка при выполнении списания средств: %v", txErr)
//			}
//			fmt.Println("responseCharge.Status", responseCharge.Status)
//			if responseCharge.Status == "CONFIRMED" {
//				fmt.Println("user.SubscribedData", user.SubscribedData)
//				subscribedDate := now.AddDate(0, 0, 1) // now.AddDate(0, 1, 0)
//				user.SubscribedData = subscribedDate.Format("2006-01-02")
//				if txErr := tx.Save(&user).Error; txErr != nil {
//					return fmt.Errorf("Ошибка при обновлении данных пользователя: %v", txErr)
//				}
//			}
//			OrderIdInt, txErr := strconv.Atoi(responseCharge.OrderId)
//			transaction := Transaction{
//				Type:        "Tinkoff",
//				Hash:        "",
//				OrderID:     uint(OrderIdInt),
//				UserID:      subscribeInit.UserID,
//				Amount:      PayPriceRUB,
//				Description: RenewalDescription,
//				Phone:       user.Phone,
//				Email:       user.Email,
//				Status:      responseCharge.Status,
//				PaymentId:   responseCharge.PaymentId,
//				Taken:       false,
//				Date:        date,
//			}
//			// Сохранить transaction в базе данных
//			if txErr := tx.Create(&transaction).Error; txErr != nil {
//				return fmt.Errorf("Ошибка при сохранении транзакции: %v", txErr)
//			}
//		}
//
//		return nil
//	})
//	if err != nil {
//		return "", fmt.Errorf("Ошибка при обработке подписки: %v", err)
//	}
//
//	return user.SubscribedData, nil
//}
