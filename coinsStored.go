package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

// Coins Stored
type CoinsStored struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Value     float32            `json:"value" bson:"value"`
	Type      string             `json:"type" bson:"type"`
	Balance   int                `json:"balance" bson:"balance"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

func initStoreCashier(cashierID string) error {
	currentTime := time.Now()
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	coinStoreCollection := cashierDatabase.Collection(CoinStoreCollectionName)
	for _, CoinsStoreInit := range CoinsStoreInits {
		CoinsStoreInitClone := make(map[string]interface{})
		CoinsStoreInitClone = CoinsStoreInit
		CoinsStoreInitClone["cashier_id"] = cashierID
		CoinsStoreInitClone["updated_at"] = currentTime
		ctx := context.Background()
		opts := options.Update().SetUpsert(true)
		filter := bson.M{"_id": primitive.NewObjectID()}
		update := bson.M{"$set": CoinsStoreInitClone}
		_, err := coinStoreCollection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return err
		}
	}
	return nil
}
func CreateCashLog(action string, auditedChanges []map[string]interface{}) error {
	currentTime := time.Now()
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	CashLogsCollection := cashierDatabase.Collection(CashLogsCollectionName)
	cashierDetailMap := map[string]interface{}{
		"_id":             primitive.NewObjectID(),
		"action":          action,
		"audited_changes": auditedChanges,
		"created_at":      currentTime,
	}
	return resource.Mongo.Insert(CashLogsCollection, cashierDetailMap)
}
