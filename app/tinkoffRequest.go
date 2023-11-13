package app

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

var InitURL string
var StateURL string
var CancelURL string
var GetCardListURL string
var ChargeURL string

func requestInit(orderId string, userID string, amount string, description string, data map[string]string, receipt map[string]interface{}) (ResponseInit, error) {

	// Собираем массив всех параметров
	values := []string{
		amount,
		description,
		orderId,
		TerminalPass,
		TerminalKey,
	}
	// Конкатенируем значения всех параметров
	dataString := strings.Join(values, "")

	// Вычисляем хэш SHA-256 для полученной строки
	hash := sha256.Sum256([]byte(dataString))
	hashString := hex.EncodeToString(hash[:])

	// Формируем токен
	token := strings.ToLower(hashString)

	// Формируем данные для запроса
	requestData := map[string]interface{}{
		"TerminalKey": TerminalKey,
		"Amount":      amount,
		"OrderId":     orderId,
		"CustomerKey": userID,
		"Description": description,
		"Token":       token,
		"SuccessURL":  SuccessURL,
		"FailURL":     FailURL,
		"DATA":        data,
		"Receipt":     receipt,
	}

	fmt.Println("requestInitURL: ", InitURL)
	fmt.Println("requestInitData: ", requestData)
	response, err := sendRequestInit(InitURL, requestData)
	fmt.Println("responseInit:", response)

	if err != nil {
		return ResponseInit{}, err
	}

	responseInit := ResponseInit{
		PaymentURL: response.PaymentURL,
		PaymentId:  response.PaymentId,
		Success:    response.Success,
		Message:    response.Message,
		Status:     response.Status,
	}

	return responseInit, nil
}

func RequestGetCardList(сustomerKey string) (ResponseGetCardList, error) {


	// Собираем массив всех параметров
	values := []string{
		сustomerKey,
		TerminalPass,
		TerminalKey,
	}
	// Конкатенируем значения всех параметров
	dataString := strings.Join(values, "")

	// Вычисляем хэш SHA-256 для полученной строки
	hash := sha256.Sum256([]byte(dataString))
	hashString := hex.EncodeToString(hash[:])

	// Формируем токен
	token := strings.ToLower(hashString)

	// Формируем данные для запроса
	requestData := map[string]interface{}{
		"TerminalKey": TerminalKey,
		"CustomerKey": сustomerKey,
		"Token":       token,
	}

	fmt.Println("requestGetCardListURL: ", GetCardListURL)
	fmt.Println("requestGetCardListData: ", requestData)
	response, err := sendRequestGetCardList(GetCardListURL, requestData)
	fmt.Println("responseInit:", response)

	if err != nil {
		return ResponseGetCardList{}, err
	}

	// Перебираем все карты в ответе
	for _, card := range response {
		// Проверяем, удовлетворяет ли карта нашим условиям
		if (card.CardType == 2 || card.CardType == 0) && card.Status == "A" {
			// Возвращаем найденную карту
			responseGetCardList := ResponseGetCardList{
				RebillId: card.RebillId,
			}
			return responseGetCardList, nil
		}
	}

	// Если ни одна карта не удовлетворяет условиям, возвращаем ошибку
	return ResponseGetCardList{}, fmt.Errorf("no suitable card found")
}

func RequestCharge(paymentId string, rebillId string) (ResponseCharge, error) {

	// Собираем массив всех параметров
	values := []string{
		TerminalPass,
		paymentId,
		rebillId,
		TerminalKey,
	}
	// Конкатенируем значения всех параметров
	dataString := strings.Join(values, "")

	// Вычисляем хэш SHA-256 для полученной строки
	hash := sha256.Sum256([]byte(dataString))
	hashString := hex.EncodeToString(hash[:])

	// Формируем токен
	token := strings.ToLower(hashString)

	requestData := map[string]interface{}{
		"TerminalKey": TerminalKey,
		"PaymentId":   paymentId,
		"RebillId":    rebillId,
		"Token":       token,
	}

	fmt.Println("requestChargeURL: ", ChargeURL)
	fmt.Println("requestChargeData: ", requestData)
	response, err := sendRequestCharge(ChargeURL, requestData)
	fmt.Println("responseState:", response)

	if err != nil {
		fmt.Println(err)
		return ResponseCharge{}, err
	}

	return ResponseCharge{
		PaymentId: response.PaymentId,
		Success:   response.Success,
		Message:   response.Message,
		Status:    response.Status,
		OrderId:   response.OrderId,
	}, nil
}

func requestSubscribeInit(orderId string, userID string, amount string, description string, data map[string]string, receipt map[string]interface{}) (ResponseInit, error) {

	// Собираем массив всех параметров
	values := []string{
		amount,
		userID,
		description,
		orderId,
		TerminalPass,
		"Y",
		TerminalKey,
	}
	// Конкатенируем значения всех параметров
	dataString := strings.Join(values, "")

	// Вычисляем хэш SHA-256 для полученной строки
	hash := sha256.Sum256([]byte(dataString))
	hashString := hex.EncodeToString(hash[:])

	// Формируем токен
	token := strings.ToLower(hashString)

	// Формируем данные для запроса
	requestData := map[string]interface{}{
		"TerminalKey": TerminalKey,
		"Amount":      amount,
		"OrderId":     orderId,
		"CustomerKey": userID,
		"Recurrent":   "Y",
		"Description": description,
		"Token":       token,
		"SuccessURL":  SuccessURL,
		"FailURL":     FailURL,
		"DATA":        data,
		"Receipt":     receipt,
	}

	fmt.Println("requestInitURL: ", InitURL)
	fmt.Println("requestInitData: ", requestData)
	response, err := sendRequestInit(InitURL, requestData)
	fmt.Println("responseInit:", response)

	if err != nil {
		return ResponseInit{}, err
	}

	responseInit := ResponseInit{
		PaymentURL: response.PaymentURL,
		PaymentId:  response.PaymentId,
		Success:    response.Success,
		Message:    response.Message,
		Status:     response.Status,
	}

	return responseInit, nil
}

func requestState(paymentId string) (ResponseState, error) {

	// Собираем массив всех параметров
	values := []string{
		TerminalPass,
		paymentId,
		TerminalKey,
	}
	// Конкатенируем значения всех параметров
	dataString := strings.Join(values, "")

	// Вычисляем хэш SHA-256 для полученной строки
	hash := sha256.Sum256([]byte(dataString))
	hashString := hex.EncodeToString(hash[:])

	// Формируем токен
	token := strings.ToLower(hashString)

	requestData := map[string]interface{}{
		"TerminalKey": TerminalKey,
		"PaymentId":   paymentId,
		"Token":       token,
	}

	fmt.Println("requestStateURL: ", StateURL)
	fmt.Println("requestStateData: ", requestData)
	response, err := sendRequestState(StateURL, requestData)
	fmt.Println("responseState:", response)

	if err != nil {
		fmt.Println(err)
		return ResponseState{}, err
	}

	return ResponseState{
		Success: response.Success,
		Status:  response.Status,
	}, nil
}

func requestCancel(paymentId string) (ResponseCancel, error) {

	// Собираем массив всех параметров
	values := []string{
		TerminalPass,
		paymentId,
		TerminalKey,
	}
	// Конкатенируем значения всех параметров
	dataString := strings.Join(values, "")

	// Вычисляем хэш SHA-256 для полученной строки
	hash := sha256.Sum256([]byte(dataString))
	hashString := hex.EncodeToString(hash[:])

	// Формируем токен
	token := strings.ToLower(hashString)

	requestData := map[string]interface{}{
		"TerminalKey": TerminalKey,
		"PaymentId":   paymentId,
		"Token":       token,
	}

	fmt.Println("requestCancelURL: ", CancelURL)
	fmt.Println("requestCancelData: ", requestData)
	response, err := sendRequestCancel(CancelURL, requestData)
	fmt.Println("responseCancel:", response)

	if err != nil {
		fmt.Println(err)
		return ResponseCancel{}, err
	}

	return ResponseCancel{
		Success:   response.Success,
		Status:    response.Status,
		PaymentId: response.PaymentId,
	}, nil
}
