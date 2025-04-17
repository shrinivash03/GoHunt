package db

import (
	"fmt"
	"os"

	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DBconn *gorm.DB

func InitDB() {
	dburl := os.Getenv("DATABASE_URL")
	var err error
	DBconn, err = gorm.Open(postgres.Open(dburl))
	if err != nil {
		fmt.Println("failed to connect to database")
		panic("failed to connect to database")
	}

	err = DBconn.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	if err != nil {
		fmt.Println("can't install uuid extension")
		panic(err)
	}
	
	err = DBconn.AutoMigrate(&User{}, &SearchSettings{}, &CrawledUrl{}, &SearchIndex{})
	if err != nil {
		fmt.Println("failed to migrate")
		panic(err)
	}
}

func GetDB() *gorm.DB {
	return DBconn
}