package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	mongolib "go.mongodb.org/mongo-driver/mongo"
)

// cashier detail
type CashierDetail struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	CashierID string             `json:"cashier_id" bson:"cashier_id"`
	Location  string             `json:"location" bson:"location"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	IsActive  bool               `json:"is_active" bson:"is_active"`
}
type ProductDetail struct {
	ID         int32   `json:"id" bson:"id"`
	Title      string  `json:"title" bson:"title"`
	Price      float32 `json:"price" bson:"price"`
	TotalPrice float32 `json:"total_price" bson:"total_price"`
}
type ProductDetails struct {
	Count          int             `json:"count" bson:"count"`
	ProductDetails []ProductDetail `json:"product_details" bson:"product_details"`
}
type OrderDetail struct {
	ProductDetails ProductDetails `json:"orders" bson:"orders"`
	TotalPrice     float32        `json:"total_price" bson:"total_pice"`
	ReceiveCash    float32        `json:"receive_cash" bson:"receive_cash"`
	Change         float32        `json:"change" bson:"change"`
}

func GetCashierDetails(collection *mongolib.Collection, cashierId string) ([]map[string]interface{}, error) {
	query := primitive.M{"cashier_id": cashierId}
	cashierDetail, err := resource.Mongo.GetAll(collection, query)
	if err != nil {
		return cashierDetail, err
	} else {
		return cashierDetail, nil
	}
}
func GetCashierDetail(collection *mongolib.Collection, cashierId string) (map[string]interface{}, error) {
	query := primitive.M{"cashier_id": cashierId}
	cashierDetail, err := resource.Mongo.GetOne(collection, query)
	if err != nil {
		return cashierDetail, err
	} else {
		return cashierDetail, nil
	}
}
func CreateCashOrder(orderDetail OrderDetail) error {
	currentTime := time.Now()
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	OrderLogsCollection := cashierDatabase.Collection(OrderLogsCollectionName)
	cashierDetailMap := map[string]interface{}{
		"_id":        primitive.NewObjectID(),
		"order":      orderDetail.ProductDetails,
		"created_at": currentTime,
	}
	return resource.Mongo.Insert(OrderLogsCollection, cashierDetailMap)
}
