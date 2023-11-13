package app

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
)

// GetProfile обрабатывает GET-запросы на /profile
func GetProfile(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.MustGet("userID").(uint)

	var user User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

    func SetupRoutes(db *gorm.DB) *gin.Engine {
        router := gin.Default()

        config := cors.DefaultConfig()
        config.AllowAllOrigins = true
        config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
        config.AllowHeaders = []string{
            "Origin",
            "Content-Length",
            "Content-Type",
            "Authorization",
        }
        config.ExposeHeaders = []string{"Content-Length"}
        config.AllowCredentials = true

        router.Use(cors.New(config))

        // Add the db to the context
        router.Use(func(c *gin.Context) {
            c.Set("db", db)
            c.Next()
        })

        router.POST("/stripe-webhook", StripeWebhook)
        v1 := router.Group("/api/v1")
        {
            v1.POST("/check-token-validity", CheckToken)
            v1.Use(AuthRequired())
            {
                v1.POST("/payment-mir", PaymentMIR)
                v1.POST("/payment-stripe", PaymentStripe)

                v1.POST("/subscribe", Subscribe)
                v1.POST("/payment-crypto", PaymentCrypto)
                v1.POST("/subscribe-crypto", SubscribeCrypto)
                v1.GET("/check-access", CheckAccess)
                v1.GET("/user-transactions-updated", UserTransactionsUpdated)
                v1.GET("/unsubscribe", UnSubscribe)

                //v1.GET("/check-subscribe", CheckSubscription)
                //v1.GET("/user-subscribe-status", UserSubscribeStatus)
                //v1.GET("/user-subscribe-status-crypto", UserSubscribeStatusCrypto)
                v1.GET("/profile", GetProfile)
                v1.Use(AdminRequired())
                {

                    v1.POST("/get-user", GetUserData)
                    v1.POST("/correct-user", ChangeUserData)
                    v1.DELETE("/payment-remove", RemovePayment)

                    //v1.POST("/user_stats", UserStats)
                    v1.POST("/get-data", GetData)
                    v1.POST("/payment-cancel", CancelPayment)

                }
            }
        }

        // Serve frontend static files
        router.StaticFS("/static", http.Dir("./frontend/build/static"))
        router.StaticFile("/", "./frontend/build/index.html")

        return router
    }
