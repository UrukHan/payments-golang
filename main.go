package main

import (
	"auth/app"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
	"os"
)

func main() {
	// Setup configuration
	viper.SetConfigFile("./config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}


	app.PayPriceMIR = viper.GetInt("payPriceMIR")
	app.SubscribePriceMIR = viper.GetInt("subscribePriceMIR")
	app.PayPriceCrypto = viper.GetInt("payPriceCrypto")
	app.SubscribePriceCrypto = viper.GetInt("subscribePriceCrypto")
	app.TestDays = viper.GetInt("testDays")
	app.AccessDays = viper.GetInt("accessDays")

	app.PolygonOwner = os.Getenv("POLYGON_OWNER")
	app.PolygonProvider = os.Getenv("POLYGON_PROVIDER")
	app.PolygonPrivateKey = os.Getenv("POLYGON_PRIVATE_KEY")
	app.PolygonID = viper.GetInt("polygonID")
	app.PolygonContract = viper.GetString("polygonContract")
	app.PolygonUSDT = viper.GetString("polygonUSDT")
	app.PolygonDecimal = viper.GetInt("polygonDecimal")

	app.AdminEmail = os.Getenv("ADMIN_EMAIL")
	app.AdminPassword = os.Getenv("ADMIN_PASSWORD")
	app.JwtSecret = os.Getenv("JWTSECRET")

	app.TerminalKey = os.Getenv("TERMINAL_KEY")
	app.TerminalPass = os.Getenv("TERMINAL_PASSWORD")

    app.StripeKeyTest = os.Getenv("STRIPE_SECRET_TEST_KEY")
    app.StripeKey = os.Getenv("STRIPE_SECRET_KEY")
    app.PaymentStripeWebhookKey = "whsec_8Pp4MeueOoANji3I9GyiyXGd5gioSwnj" //os.Getenv("PAY_STRIPE_WEBHOOK")

	app.PaymentDescription = viper.GetString("paymentDescription")
	app.SubscribeDescription = viper.GetString("subscribeDescription")
	app.RenewalDescription = viper.GetString("renewalDescription")
	app.RenewalInitDescription = viper.GetString("renewalInitDescription")

	app.StateURL = viper.GetString("getStateURL")
	app.InitURL = viper.GetString("initTxURL")
	app.CancelURL = viper.GetString("cancelPaymentURL")
	app.GetCardListURL = viper.GetString("getCardListURL")
	app.ChargeURL = viper.GetString("chargeURL")
	app.SuccessURL = viper.GetString("successURL")
	app.FailURL = viper.GetString("failURL")

	// Connect to PostgreSQL
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_DBNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_SSLMODE")))
	if err != nil {
		fmt.Println(err)
		panic("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&app.User{})
	db.AutoMigrate(&app.Transaction{})

	// Adding routes to the same router
	r := app.SetupRoutes(db)

	// Run auto subscriber function in goroutine
	go app.AutoRenewSubscriptions(db)

	// Run un 0.0.0.0:8010
	r.Run(":8010")
}
