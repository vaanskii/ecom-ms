package db

import "gorm.io/gorm"

var db *gorm.DB

type Product struct {
	ID 		string 	`gorm:"primaryKey"`
	Name 	string
	Price 	float32
}

func GetProductByID(id string) (*Product, error) {
	var product Product

	result := db.First(&product, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &product, nil
}

func GetAllProducts() ([]Product, error) {
	var products []Product

	result := db.Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}

	return products, nil
}