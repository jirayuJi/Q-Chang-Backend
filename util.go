package main

import (
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongolib "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

// ValidatePageAndLimit TODO
func ValidatePageAndLimit(page string, limit string) (int, int) {
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 1
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10
	}
	return pageInt, limitInt
}

// GetAllWithPageLimitAndHint is find all data in collection with page and limit
func GetAllWithPageLimitAndSort(collection *mongolib.Collection, query map[string]interface{}, page int, limit int, sort string) ([]map[string]interface{}, error) {
	var (
		jsonDocuments []map[string]interface{}
		err           error
	)
	ctx := context.Background()
	findOptions := options.Find()
	if page == 0 && limit == 0 {
		findOptions.SetSort(bson.D{{sort, -1}})
	} else {
		findOptions.SetSort(bson.D{{sort, -1}}).SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	}
	cursor, _ := collection.Find(nil, query, findOptions)
	if err = cursor.All(ctx, &jsonDocuments); err != nil {
		return make([]map[string]interface{}, 0), err
	}
	return jsonDocuments, nil
}
func GetAllWithSort(collection *mongolib.Collection, query map[string]interface{}, sort string) ([]map[string]interface{}, error) {
	var (
		jsonDocuments []map[string]interface{}
		err           error
	)
	ctx := context.Background()
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{sort, -1}})
	cursor, _ := collection.Find(nil, query, findOptions)
	if err = cursor.All(ctx, &jsonDocuments); err != nil {
		return make([]map[string]interface{}, 0), err
	}
	return jsonDocuments, nil
}
func Cloner(data map[string]interface{}) map[string]interface{} {
	clone := make(map[string]interface{})
	for key, value := range data {
		switch value.(type) {
		case map[string]interface{}:
			clone[key] = Cloner(value.(map[string]interface{}))

		default:
			clone[key] = value

		}
	}
	return clone
}
func (resource Resource) cacherProduct() {
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	productCollection := cashierDatabase.Collection(ProductCollectionName)
	query := primitive.M{}
	cashierDetails, err := resource.Mongo.GetAll(productCollection, query)
	if err == nil {
		for _, cashierDetail := range cashierDetails {
			idProductStr := strconv.Itoa(int(cashierDetail["id"].(int32)))
			ProductDetials[idProductStr] = cashierDetail
		}
	} else {
		fmt.Println("Product cacher error :", err.Error())
	}
	// fmt.Println("Product cacher done. data :", ProductDetials)
}

func convertToFloat32(data interface{}) float32 {
	var value float32
	switch data.(type) {
	case float64:
		if dataFloat64, ok := data.(float64); ok {
			return float32(dataFloat64)
		}
	case int32:
		if dataInt32, ok := data.(int32); ok {
			return float32(dataInt32)
		}
	case int:
		if dataInt32, ok := data.(int); ok {
			return float32(dataInt32)
		}
	case string:
		if dataStr, ok := data.(string); ok {
			dataFloat64, err := strconv.ParseFloat(dataStr, 32)
			if err != nil {
				return value
			}
			return float32(dataFloat64)
		}
	}
	return value
}
