package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error){
	dsn := os.Getenv("DB_URL")
	if dsn == ""{
		return nil, fmt.Errorf("db url not specified");
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{});

	if err != nil{
		return nil, fmt.Errorf("error connecting to db");
	}
	
	log.Println("server connected to db")
	return db, nil;
}
