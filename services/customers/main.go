package main

import (
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/dtm/dtmutil"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"net/http"
	"os"
)

// model
type (
	Customer struct {
		gorm.Model
		Balance uint
		Version uint
	}
	ProcessedTransaction struct {
		gorm.Model
		IDTransaction string
	}
)

// system
func main() {
	app := gin.New()

	// public customers api
	app.POST("/create", func(c *gin.Context) {
		customer := Customer{Balance: 100}

		err := getDb().Save(&customer).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, customer)
	})

	// internal customers api
	app.POST("/withdraw-money", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		withdrawRequest := struct {
			IdCustomer uint `json:"idCustomer"`
			Amount     uint `json:"amount"`
		}{}
		transactionId := c.Query("gid")

		err := c.BindJSON(&withdrawRequest)
		if err != nil {
			return dtmcli.ErrFailure
		}

		// transaction
		err = getDb().Transaction(func(tx *gorm.DB) error {
			// find customer
			var customer Customer
			err = tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&customer, withdrawRequest.IdCustomer).
				Error
			if err != nil {
				return dtmcli.ErrFailure
			}

			// check balance
			if customer.Balance < withdrawRequest.Amount {
				return dtmcli.ErrFailure
			}

			// change
			customer.Version = customer.Version + 1
			customer.Balance -= withdrawRequest.Amount

			log.Printf("change %v", customer)

			// save
			res := tx.Model(&Customer{}).
				Where("id = ? AND version = ?", customer.ID, customer.Version-1).
				Save(customer)
			if res.RowsAffected != 1 {
				return dtmcli.ErrOngoing
			}
			tx.Save(&ProcessedTransaction{IDTransaction: transactionId})

			return nil
		})

		log.Println("change result", err)

		return err
	}))
	app.POST("/withdraw-money-compensate", dtmutil.WrapHandler2(func(c *gin.Context) interface{} {
		compensateRequest := struct {
			IdCustomer uint `json:"idCustomer"`
			Amount     uint `json:"amount"`
		}{}
		transactionId := c.Query("gid")

		err := c.BindJSON(&compensateRequest)
		if err != nil {
			return dtmcli.ErrFailure
		}

		// transaction
		err = getDb().Transaction(func(tx *gorm.DB) error {
			// filter regular cases
			var pt ProcessedTransaction
			err = tx.Where(&ProcessedTransaction{IDTransaction: transactionId}).First(&pt).Error
			if err == gorm.ErrRecordNotFound {
				return nil
			}

			// change customer
			var customer Customer
			err = tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&customer, compensateRequest.IdCustomer).
				Error
			if err != nil {
				return dtmcli.ErrFailure
			}

			customer.Version = customer.Version + 1
			customer.Balance = customer.Balance + compensateRequest.Amount

			// save
			result := tx.
				Model(&Customer{}).
				Where("id = ? AND version = ?", customer.ID, customer.Version-1).
				Updates(customer)
			if result.RowsAffected == 0 {
				return dtmcli.ErrOngoing
			}

			return nil
		})

		return err
	}))

	log.Println("started")
	_ = app.Run(":8080")
}

func getDb() *gorm.DB {
	db, err := gorm.Open(mysql.Open(os.Getenv("MYSQL_DSN")), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	_ = db.AutoMigrate(&Customer{})
	_ = db.AutoMigrate(&ProcessedTransaction{})

	return db
}

func init() {}
