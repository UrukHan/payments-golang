package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type ResponseInit struct {
	Success    bool
	Message    string
	Status     string
	PaymentId  string
	PaymentURL string
}

type ResponseState struct {
	Success     bool
	ErrorCode   string
	Message     string
	TerminalKey string
	Status      string
	PaymentId   string
	OrderId     string
	Amount      int
}

type ResponseCancel struct {
	Success   bool
	Status    string
	PaymentId string
}

type ResponseCharge struct {
	Success     bool
	ErrorCode   string
	Message     string
	TerminalKey string
	Status      string
	PaymentId   string
	OrderId     string
}

type ResponseGetCardList struct {
	CardID   string `json:"CardId"`
	Pan      string `json:"Pan"`
	ExpDate  string `json:"ExpDate"`
	CardType int    `json:"CardType"`
	Status   string `json:"Status"`
	RebillId string `json:"RebillId,omitempty"`
}

func sendRequestCharge(url string, requestData map[string]interface{}) (ResponseCharge, error) {
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return ResponseCharge{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return ResponseCharge{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ResponseCharge{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ResponseCharge{}, err
	}

	var responseCharge ResponseCharge
	err = json.Unmarshal(body, &responseCharge)
	if err != nil {
		return ResponseCharge{}, err
	}

	return responseCharge, nil
}

func sendRequestGetCardList(url string, requestData map[string]interface{}) ([]ResponseGetCardList, error) {
	jsonReq, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonReq))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cards []ResponseGetCardList
	err = json.Unmarshal(body, &cards)
	if err != nil {
		return nil, err
	}

	return cards, nil
}

func sendRequestInit(url string, requestData map[string]interface{}) (ResponseInit, error) {
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return ResponseInit{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return ResponseInit{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ResponseInit{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ResponseInit{}, err
	}

	var responseInit ResponseInit
	err = json.Unmarshal(body, &responseInit)
	if err != nil {
		return ResponseInit{}, err
	}

	return responseInit, nil
}

func sendRequestState(url string, requestData map[string]interface{}) (ResponseState, error) {
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return ResponseState{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return ResponseState{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ResponseState{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ResponseState{}, err
	}

	var responseState ResponseState
	err = json.Unmarshal(body, &responseState)
	if err != nil {
		return ResponseState{}, err
	}

	return responseState, nil
}

func sendRequestCancel(url string, requestData map[string]interface{}) (ResponseCancel, error) {
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return ResponseCancel{}, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return ResponseCancel{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ResponseCancel{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ResponseCancel{}, err
	}

	var responseCancel ResponseCancel
	err = json.Unmarshal(body, &responseCancel)
	if err != nil {
		return ResponseCancel{}, err
	}

	return responseCancel, nil
}
