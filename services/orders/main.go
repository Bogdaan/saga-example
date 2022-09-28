package main

import (
	"encoding/json"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
)

// config
var dtmCoordinatorAddress = os.Getenv("DTM_COORDINATOR")
var ordersServerURL = os.Getenv("ORDERS_SERVICE_URL")
var customersServerURL = os.Getenv("CUSTOMERS_SERVICE_URL")

// model
type Order struct {
	gorm.Model
	IDTransaction string
	IDCustomer    uint
	Currency      string
	Amount        uint
	Status        string
}

// system
func main() {
	app := gin.New()

	// public order api
	app.POST("/create", func(c *gin.Context) {
		createOrderRequest := struct {
			IdCustomer uint   `json:"idCustomer"`
			Currency   string `json:"currency"`
			Amount     uint   `json:"amount"`
		}{}

		err := c.BindJSON(&createOrderRequest)
		if err != nil {
			c.JSON(extractCode(err), err)
			return
		}

		globalTransactionId := dtmcli.MustGenGid(dtmCoordinatorAddress)
		req, _ := StructToMap(createOrderRequest)
		err = dtmcli.
			NewSaga(dtmCoordinatorAddress, globalTransactionId).
			Add(ordersServerURL+"/register-order", ordersServerURL+"/register-order-compensate", req).
			Add(customersServerURL+"/withdraw-money", customersServerURL+"/withdraw-money-compensate", req).
			Submit()

		createOrderResponse := struct {
			Gid string `json:"gid"`
		}{Gid: globalTransactionId}

		c.JSON(extractCode(err), createOrderResponse)
	})

	// internal order api
	app.POST("/register-order", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		registerOrderRequest := struct {
			IdCustomer uint   `json:"idCustomer"`
			Currency   string `json:"currency"`
			Amount     uint   `json:"amount"`
		}{}
		transactionId := c.Query("gid")

		err := c.BindJSON(&registerOrderRequest)
		if err != nil {
			return err
		}

		return getDb().
			Create(&Order{
				IDTransaction: transactionId,
				IDCustomer:    registerOrderRequest.IdCustomer,
				Currency:      registerOrderRequest.Currency,
				Amount:        registerOrderRequest.Amount,
				Status:        "created",
			}).
			Error
	}))
	app.POST("/register-order-compensate", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		transactionId := c.Query("gid")

		return getDb().
			Model(&Order{}).
			Where("id_transaction = ?", transactionId).
			Update("status", "canceled").
			Limit(1).
			Error
	}))

	log.Println("started")
	_ = app.Run(":8080")
}

func extractCode(err error) int {
	statusCode := http.StatusOK
	if err != nil {
		statusCode = http.StatusInternalServerError
	}
	return statusCode
}

func getDb() *gorm.DB {
	db, err := gorm.Open(mysql.Open(os.Getenv("MYSQL_DSN")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	_ = db.AutoMigrate(&Order{})

	return db
}

func StructToMap(obj interface{}) (newMap map[string]interface{}, err error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &newMap)
	return
}
