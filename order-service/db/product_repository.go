package db

import (
	"time"

	"gorm.io/gorm"
)

var db *gorm.DB

type Orders struct {
	OrderID        string      `gorm:"primaryKey"`
	ProductID      string      `gorm:"primaryKey"`
	CustomerName   string
	Quantity       int32
	Status         string
	CreatedAt      time.Time  
}

func GetDBInstance() *gorm.DB {
    return db
}