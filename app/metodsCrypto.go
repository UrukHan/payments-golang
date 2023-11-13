package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"math/big"
	"strings"
	"time"
)

func CryptoPaymentPolygon(c *gin.Context, txHash string, txType string) (gin.H, error) {
	now := time.Now()
	date := now.Format("2006-01-02")

	db := c.MustGet("db").(*gorm.DB)
	userID, _ := GetUserC(c)

	orderID, err := GetNewOrder(db)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении последнего заказа: %v", err)
	}

	// Создание новой записи в таблице Crypto
	crypto := Transaction{
		UserID:          userID,
		Type:            txType,
		TransactionType: "Pay",
		OrderID:         orderID,
		PaymentId:       txHash,
		Status:          "pending",
		Date:            date,
		Taken:           false,
	}
	// Сохранить transaction в базе данных
	result := db.Create(&crypto)

	if result.Error != nil {
		return nil, fmt.Errorf("Ошибка при сохранении подписки: %v", result.Error)
	}

	data := gin.H{
		"paymentStatus": crypto.Status,
		"success":       true,
		"date":          date,
	}

	return data, nil
}

func CryptoSubscribePolygon(c *gin.Context, txHash string, txType string) (gin.H, error) {
	now := time.Now()
	date := now.Format("2006-01-02")

	db := c.MustGet("db").(*gorm.DB)
	userID, _ := GetUserC(c)

	orderID, err := GetNewOrder(db)
	if err != nil {
		return nil, fmt.Errorf("Ошибка при получении последнего заказа: %v", err)
	}

	subscription := Transaction{
		UserID:          userID,
		Type:            txType,
		TransactionType: "Subscribe",
		OrderID:         orderID,
		PaymentId:       txHash,
		Status:          "pending",
		Date:            date,
		Taken:           false,
	}
	result := db.Create(&subscription)
	if result.Error != nil {
		return nil, fmt.Errorf("Failed to save subscription: %v", result.Error)
	}

	data := gin.H{
		"paymentStatus": subscription.Status,
		"success":       true,
		"date":          date,
	}
	return data, nil
}

func UpdateTransactionEVM(transaction *Transaction, db *gorm.DB, user *User) error {

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var txState string
	var err error
	switch transaction.Type {
	case "USDT polygon":
		txState, err = checkTransactionEVM(PolygonProvider, transaction.PaymentId)
	default:
		tx.Rollback()
		return errors.New("Unknown transaction type")
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	if txState == "failed" {
		transaction.Status = "failed"
		transaction.Taken = true
		if err := tx.Save(transaction).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else if txState == "pending" {
		return errors.New("transaction pending")
	} else if txState == "success" {
		if transaction.TransactionType == "Pay" {
			var amount int
			txState, amount, err = CheckStatusPayEVM(PolygonProvider, transaction.PaymentId, PolygonContract, PolygonDecimal)
			if err != nil {
				return err
			}
			if txState == "success" {
				transaction.Status = "success"
				transaction.Taken = true
				transaction.Amount = amount
				transaction.Email = user.Email
				if err := tx.Save(transaction).Error; err != nil {
					tx.Rollback()
					return err
				}
				now := time.Now()
				userAccessTo := user.AccessTo
				if userAccessTo.Before(now) {
					user.AccessTo = now.AddDate(0, 0, AccessDays)
				} else {
					user.AccessTo = userAccessTo.AddDate(0, 0, AccessDays)
				}
				if err := tx.Save(user).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		} else if transaction.TransactionType == "SubscribePay" {
			var amount int
			txState, amount, err = CheckStatusPayOperatorEVM(PolygonProvider, transaction.PaymentId, PolygonContract, PolygonDecimal)
			if err != nil {
				return err
			}
			if txState == "success" {
				transaction.Status = "success"
				transaction.Taken = true
				transaction.Amount = amount
				transaction.Email = user.Email
				if err := tx.Save(transaction).Error; err != nil {
					tx.Rollback()
					return err
				}
				now := time.Now()
				userAccessTo := user.AccessTo
				if userAccessTo.Before(now) {
					user.AccessTo = now.AddDate(0, 0, AccessDays)
				} else {
					user.AccessTo = userAccessTo.AddDate(0, 0, AccessDays)
				}
				if err := tx.Save(user).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		} else if transaction.TransactionType == "Subscribe" {
			txState, err = CheckCryptoSubscriptionEVM(PolygonProvider, transaction.PaymentId, PolygonUSDT, PolygonContract, PolygonDecimal)
			if err != nil {
				errMsg := fmt.Sprintf("UpdateTransactionEVM: CheckCryptoSubscriptionEVM %s", err)
				fmt.Println(errMsg)
				tx.Rollback()
				return err
			}
			if txState == "success" {
				transaction.Status = "success"
				transaction.Taken = true
				transaction.Email = user.Email
				if err := tx.Save(transaction).Error; err != nil {
					tx.Rollback()
					return err
				}
				now := time.Now()
				if user.AccessTo.IsZero() {
					var userAccessTo time.Time
					if !user.AccessTo.IsZero() {
						userAccessTo = user.AccessTo
					}
					if userAccessTo.Before(now) {
						user.AccessTo = now.AddDate(0, 0, TestDays)
					} else {
						user.AccessTo = userAccessTo.AddDate(0, 0, TestDays)
					}
				} else {
					switch transaction.Type {
					case "USDT polygon", "USDC polygon":
						AutoPaymentEVM(PolygonID, PolygonContract, PolygonPrivateKey, PolygonProvider, user, db)
					}
				}
				user.Subscribe = transaction.Type
				fmt.Println("user", user)
				if err := tx.Save(user).Error; err != nil {
					tx.Rollback()
					return err
				}
			} else if txState == "failed" {
				transaction.Status = "failed"
				transaction.Taken = true
				if err := tx.Save(transaction).Error; err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	return tx.Commit().Error
}

func CheckStatusPayEVM(provider string, hash string, contract string, decimal int) (string, int, error) {

	client, err := ethclient.Dial(provider)
	if err != nil {
		fmt.Println("CheckStatusPayEVM: ethclient.Dial(provider)")
		return "", 0, err
	}

	txHash := common.HexToHash(hash)
	tx, _, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		errMsg := fmt.Sprintf("CheckStatusPayEVM: TransactionByHash %s", err)
		fmt.Println(errMsg)
		return "pending", 0, nil
	}

	contractAddress := common.HexToAddress(contract)
	isPayCall, err := isPayFunctionCall(client, txHash, contractAddress)
	if err != nil {
		errMsg := fmt.Sprintf("CheckStatusPayEVM: isPayFunctionCall %s", err)
		fmt.Println(errMsg)
		return "pending", 0, nil
	}
	if isPayCall {
		price, err := getPriceFromLogs(client, tx, contractAddress)
		if err != nil {
			errMsg := fmt.Sprintf("CheckStatusPayEVM: getPriceFromLogs %s", err)
			fmt.Println(errMsg)
			return "pending", 0, nil
		}
		price = new(big.Int).Div(price, big.NewInt(int64(decimal)))
		return "success", int(price.Int64()), nil
	}
	return "pending", 0, nil
}

func CheckStatusPayOperatorEVM(provider string, hash string, contract string, decimal int) (string, int, error) {

	client, err := ethclient.Dial(provider)
	if err != nil {
		fmt.Println("CheckStatusPayOperatorEVM: ethclient.Dial(provider)")
		return "", 0, err
	}

	txHash := common.HexToHash(hash)
	tx, _, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		errMsg := fmt.Sprintf("CheckStatusPayOperatorEVM: TransactionByHash %s", err)
		fmt.Println(errMsg)
		return "pending", 0, nil
	}

	contractAddress := common.HexToAddress(contract)
	isPayOperatorCall, err := isPayOperatorFunctionCall(client, txHash, contractAddress)
	if err != nil {
		errMsg := fmt.Sprintf("CheckStatusPayOperatorEVM: isPayOperatorFunctionCall %s", err)
		fmt.Println(errMsg)
		return "pending", 0, nil
	}
	if isPayOperatorCall {
		price, err := getPriceFromLogs(client, tx, contractAddress)
		fmt.Println("TX price:      ", price, err)
		if err != nil {
			errMsg := fmt.Sprintf("CheckStatusPayOperatorEVM: getPriceFromLogs %s", err)
			fmt.Println(errMsg)
			return "pending", 0, nil
		}
		price = new(big.Int).Div(price, big.NewInt(int64(decimal)))
		return "success", int(price.Int64()), nil
	}
	return "pending", 0, nil
}

func CheckCryptoSubscriptionEVM(provider string, hash string, currency string, contract string, decimal int) (string, error) {

	client, err := ethclient.Dial(provider)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to the Ethereum client: %v", err)
	}

	txHash := common.HexToHash(hash)

	tokenContractAddress := common.HexToAddress(currency)
	spenderAddress := common.HexToAddress(contract)
	expectedAmount := new(big.Int).Mul(big.NewInt(int64(SubscribePriceCrypto)), big.NewInt(int64(decimal)))
	isApproveCall, approvedAmount, err := isApproveFunctionCall(client, txHash, tokenContractAddress, spenderAddress)
	if err != nil || !isApproveCall {
		return "pending", nil
	}

	if approvedAmount.Cmp(expectedAmount) != 0 {
		return "failed", nil
	}

	return "success", nil
}

func AutoPaymentEVM(chain int, contract string, privateKey string, provider string, user *User, db *gorm.DB) {

	now := time.Now()
	date := now.Format("2006-01-02")

	client, err := ethclient.Dial(provider)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to connect to the Ethereum client: %s", err)
		fmt.Println(errMsg)
		return
	}

	hexPrivateKey := privateKey
	privateKeyBytes := common.FromHex(hexPrivateKey)

	privateKeyECDSA, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		fmt.Printf("Не удалось преобразовать приватный ключ: %s", err)
		return
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKeyECDSA, big.NewInt(int64(chain)))
	if err != nil {
		fmt.Printf("Не удалось создать авторизованный транзактор: %s", err)
		return
	}

	contractAddress := common.HexToAddress(contract)

	jsonAbi, err := ioutil.ReadFile("utils/abiPay.json")
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read ABI: %s", err)
		fmt.Println(errMsg)
		return
	}

	parsedAbi, err := abi.JSON(strings.NewReader(string(jsonAbi)))
	if err != nil {
		errMsg := fmt.Sprintf("Failed to parse ABI: %s", err)
		fmt.Println(errMsg)
		return
	}

	contractCollect := bind.NewBoundContract(contractAddress, parsedAbi, client, client, client)

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		errMsg := fmt.Sprintf("Failed Begin transaction: %s", tx.Error)
		fmt.Println(errMsg)
		return
	}

	var txSend *types.Transaction

	switch user.Subscribe {
	case "USDT polygon":
		txSend, err = contractCollect.Transact(auth, "payOperator", "USDT", common.HexToAddress(user.Address))
	case "USDC polygon":
		txSend, err = contractCollect.Transact(auth, "payOperator", "USDC", common.HexToAddress(user.Address))
	}

	if err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed Transact transaction: %s", err)
		fmt.Println(errMsg)
		return
	}

	orderID, err := GetNewOrder(db)
	if err != nil {
		errMsg := fmt.Sprintf("Ошибка при получении последнего заказа: %s", err)
		fmt.Println(errMsg)
		return
	}

	crypto := Transaction{
		UserID:          user.ID,
		Type:            user.Subscribe,
		TransactionType: "SubscribePay",
		OrderID:         orderID,
		PaymentId:       txSend.Hash().Hex(),
		Status:          "pending",
		Date:            date,
		Taken:           false,
	}

	result := tx.Create(&crypto)
	if result.Error != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Ошибка при получении последнего заказа: %s", result.Error)
		fmt.Println(errMsg)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		errMsg := fmt.Sprintf("Failed to commit transaction: %s", err)
		fmt.Println(errMsg)
		return
	}
}

func checkTransactionEVM(provider string, hash string) (string, error) {
	client, err := ethclient.Dial(provider)
	if err != nil {
		return "", fmt.Errorf("Failed to connect to the Ethereum client: %v", err)
	}

	txHash := common.HexToHash(hash)

	_, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve transaction by hash: %v", err)
	}

	if isPending {
		return "pending", nil
	}

	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve transaction receipt: %v", err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		return "success", nil
	} else {
		return "failed", nil
	}
}

func isPayOperatorFunctionCall(client *ethclient.Client, txHash common.Hash, contractAddress common.Address) (bool, error) {
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return false, err
	}

	if isPending {
		return false, nil
	}

	to := tx.To()
	if to == nil || *to != contractAddress {
		return false, nil
	}

	data := tx.Data()
	funcSig := []byte("payOperator(string,address)")
	hash := crypto.Keccak256(funcSig)
	funcID := hash[:4]

	return bytes.Equal(data[:4], funcID), nil
}

func isPayFunctionCall(client *ethclient.Client, txHash common.Hash, contractAddress common.Address) (bool, error) {
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return false, err
	}

	if isPending {
		return false, nil
	}

	to := tx.To()
	if to == nil || *to != contractAddress {
		return false, nil
	}

	data := tx.Data()
	funcSig := []byte("pay(string)")
	hash := crypto.Keccak256(funcSig)
	funcID := hash[:4]

	return bytes.Equal(data[:4], funcID), nil
}

func getPriceFromLogs(client *ethclient.Client, tx *types.Transaction, contractAddress common.Address) (*big.Int, error) {
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return nil, err
	}

	// Load and parse the contract's ABI
	jsonAbi, err := ioutil.ReadFile("utils/abiPay.json")
	if err != nil {
		errMsg := fmt.Sprintf("getPriceFromLogs: jsonAbi %s", err)
		fmt.Println(errMsg)
		return nil, fmt.Errorf("Failed to parse ABI: %v", err)
	}

	for _, log := range receipt.Logs {
		if log.Address == contractAddress {
			event := struct {
				Currency string
				From     common.Address
				Amount   *big.Int
			}{}

			parsedAbi, err := abi.JSON(strings.NewReader(string(jsonAbi)))
			if err != nil {
				errMsg := fmt.Sprintf("getPriceFromLogs: parsedAbi %s", err)
				fmt.Println(errMsg)
				return nil, fmt.Errorf("failed to parse ABI: %v", err)
			}

			err = parsedAbi.UnpackIntoInterface(&event, "PaymentReceived", log.Data)
			if err != nil {
				errMsg := fmt.Sprintf("getPriceFromLogs: UnpackIntoInterface %s", err)
				fmt.Println(errMsg)
				return nil, err
			}

			return event.Amount, nil
		}
	}

	return nil, errors.New("event not found in the transaction logs")
}

func isApproveFunctionCall(client *ethclient.Client, txHash common.Hash, tokenContractAddress, spenderAddress common.Address) (bool, *big.Int, error) {
	// Чтение и парсинг ABI
	bytes, err := ioutil.ReadFile("./utils/abiUSDT.json")
	if err != nil {
		return false, nil, err
	}

	contractAbi, err := abi.JSON(strings.NewReader(string(bytes)))
	if err != nil {
		return false, nil, err
	}

	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return false, nil, err
	}

	if isPending {
		return false, nil, nil
	}

	to := tx.To()
	data := tx.Data()

	// If the data is shorter than the length of the method ID, return error.
	if len(data) < 4 {
		return false, nil, fmt.Errorf("transaction data is too short")
	}

	// Get the method ID of the "approve" method.
	methodID := contractAbi.Methods["approve"].ID

	// Check if the method ID from the transaction data matches the "approve" method ID.
	if string(data[:4]) != string(methodID[:]) {
		return false, nil, fmt.Errorf("not an approve method call")
	}

	// Unpack the transaction data into an array of interfaces.
	args, err := contractAbi.Methods["approve"].Inputs.UnpackValues(data[4:])
	if err != nil {
		return false, nil, err
	}
	// Extract the spender address and amount from the unpacked arguments.
	spender := args[0].(common.Address)
	amount := args[1].(*big.Int)
	// Check if the call is to the correct contract and spender.
	isApproveCall := (to != nil && *to == tokenContractAddress && spender == spenderAddress)

	return isApproveCall, amount, nil
}
