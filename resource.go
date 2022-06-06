package main

import (
	"encoding/json"
	"fmt"
	"math"
	"mongodriver"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongolib "go.mongodb.org/mongo-driver/mongo"
)

// Attachment todo
type Attachment struct {
	Image []string `json:"image" bson:"image"`
}

// Resource TODO
type Resource struct {
	Mongo mongodriver.Mongo
}

func (resource Resource) healthcheck(c echo.Context) error {
	results := bson.M{"message": "Welcome to auto cashier system."}
	return c.JSON(http.StatusOK, results)
}

// CreateCashier TODO
func (resource Resource) CreateCashier(c echo.Context) error {
	cashierId := c.FormValue("cashier_id")
	location := c.FormValue("location")
	isActive := c.FormValue("is_active")
	isActiveBool := false
	if isActive != "" {
		isActiveInput, err := strconv.ParseBool(isActive)
		if err == nil {
			isActiveBool = isActiveInput
		} else {
			return c.JSON(http.StatusInternalServerError, bson.M{"message": "parameter is_active invalid."})
		}
	}
	if cashierId == "" || location == "" {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": "Missing required parameter (cashier_id,location)"})
	}
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	cashierCollection := cashierDatabase.Collection(CashierCollectionName)

	cashierDetial, err := GetCashierDetails(cashierCollection, cashierId)
	if err == mongolib.ErrNoDocuments || len(cashierDetial) == 0 {
		currentTime := time.Now()
		cashierDetailMap := map[string]interface{}{
			"_id":        primitive.NewObjectID(),
			"cashier_id": cashierId,
			"location":   location,
			"is_active":  isActiveBool,
			"created_at": currentTime,
			"updated_at": currentTime,
		}
		err := resource.Mongo.Insert(cashierCollection, cashierDetailMap)
		if err == nil {
			cashierDetial, err := GetCashierDetails(cashierCollection, cashierId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Cannot query cashier_id because %s", err.Error())})
			}
			dataRespons := map[string]interface{}{
				"count": len(cashierDetial),
				"data":  cashierDetial,
			}
			if err := initStoreCashier(cashierDetailMap["cashier_id"].(string)); err != nil {
				return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Cannot init store cashier error ", err)})
			}
			return c.JSON(http.StatusOK, dataRespons)
		} else {
			return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Cannot create cashier because %s", err.Error())})
		}
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Cannot query cashier_id because %s", err.Error())})
	} else {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("cashier_id %s is exist", cashierId)})
	}
}

// UpdateCashier TODO
func (resource Resource) UpdateCashier(c echo.Context) error {
	cashierId := c.FormValue("cashier_id")
	location := c.FormValue("location")
	isActive := c.FormValue("is_active")
	if cashierId == "" {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": "Missing required parameter (cashier_id)"})
	}
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	cashierCollection := cashierDatabase.Collection(CashierCollectionName)

	cashierDetail, err := GetCashierDetail(cashierCollection, cashierId)
	if err == nil || len(cashierDetail) > 0 {
		ObjIDCashierString := cashierDetail["_id"].(primitive.ObjectID).Hex()
		ObjIDCashier, err := primitive.ObjectIDFromHex(ObjIDCashierString)
		currentTime := time.Now()
		if location != "" {
			cashierDetail["location"] = location
		}
		if isActive != "" {
			isActiveInput, err := strconv.ParseBool(isActive)
			if err == nil {
				cashierDetail["is_active"] = isActiveInput
			} else {
				return c.JSON(http.StatusInternalServerError, bson.M{"message": "parameter is_active invalid."})
			}
		}
		cashierDetail["updated_at"] = currentTime
		cashierDetail["created_at"] = cashierDetail["created_at"].(primitive.DateTime).Time()
		fmt.Println("ObjIDCashier:", ObjIDCashier)
		query := primitive.M{"_id": ObjIDCashier}

		cashierDetial, err := resource.Mongo.UpdateOne(cashierCollection, query, primitive.M{"$set": cashierDetail})
		if err == nil {
			dataRespons := map[string]interface{}{
				"count": len(cashierDetial),
				"data":  cashierDetial,
			}
			return c.JSON(http.StatusOK, dataRespons)
		} else {
			return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Cannot update cashier because %s", err.Error())})
		}
	} else {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("cashier_id %s not found", cashierId)})
	}
}

func (resource Resource) GetCashiers(c echo.Context) error {
	page := c.FormValue("page")
	limit := c.FormValue("limit")
	isActive := c.FormValue("is_active")
	sortedBy := c.FormValue("sorted_by")
	pageInt, limitInt := ValidatePageAndLimit(page, limit)
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	cashierCollection := cashierDatabase.Collection(CashierCollectionName)
	sortedByQuery := "created_at"
	query := map[string]interface{}{}
	if sortedBy != "" {
		sortedByQuery = sortedBy
	}
	if isActive != "" {
		isActiveInput, err := strconv.ParseBool(isActive)
		if err == nil {
			query["is_active"] = isActiveInput
		} else {
			return c.JSON(http.StatusInternalServerError, bson.M{"message": "parameter is_active invalid."})
		}
	}

	cashierDetails, err := GetAllWithPageLimitAndSort(cashierCollection, query, pageInt, limitInt, sortedByQuery)
	if err == nil {
		dataRespons := map[string]interface{}{
			"count": len(cashierDetails),
			"data":  cashierDetails,
		}
		return c.JSON(http.StatusOK, dataRespons)
	}
	return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Something went wrong error %s", err)})
}

func (resource Resource) TopUp(c echo.Context) error {
	cashierId := c.FormValue("cashier_id")
	balance := c.FormValue("balance")
	cashType := c.FormValue("type")
	value := c.FormValue("value")
	if cashierId == "" {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": "Missing required parameter (cashier_id)"})
	}
	if balance == "" || cashType == "" || value == "" {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": "Missing required parameter (type,value,balance)"})
	}
	var cashierDetials []map[string]interface{}
	balances := strings.Split(balance, ",")
	cashTypes := strings.Split(cashType, ",")
	values := strings.Split(value, ",")
	currentTime := time.Now()
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	coinStoreCollection := cashierDatabase.Collection(CoinStoreCollectionName)
	if (len(balances) == len(cashTypes)) && (len(balances) == len(values)) {
		for index, value := range values {
			valueNew, _ := strconv.Atoi(value)
			query := map[string]interface{}{
				"value":      valueNew,
				"type":       cashTypes[index],
				"cashier_id": cashierId,
			}
			cashierDetail, err := resource.Mongo.GetOne(coinStoreCollection, query)
			cashierDetailClone := Cloner(cashierDetail)
			if err == mongolib.ErrNoDocuments || cashierDetail == nil {
				return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("cashier_id %v value %v type %v", cashierId, value, cashTypes[index])})
			} else if err != nil {
				return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Cannot get coin store because", err)})
			} else {
				balanceNew, _ := strconv.Atoi(balances[index])
				cashierDetail["balance"] = cashierDetail["balance"].(int32) + int32(balanceNew)
				cashierDetail["updated_at"] = currentTime
				cashierDetailNew, err := resource.Mongo.UpdateOne(coinStoreCollection, query, primitive.M{"$set": cashierDetail})
				if err == nil {
					auditedChanges := []map[string]interface{}{
						cashierDetailClone,
						cashierDetailNew,
					}
					CreateCashLog(actionTopUp, auditedChanges)
					cashierDetials = append(cashierDetials, cashierDetailNew)
				} else {
					return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Cannot update cashier because %s", err.Error())})
				}
			}
		}
		dataRespons := map[string]interface{}{
			"count": len(cashierDetials),
			"data":  cashierDetials,
		}
		return c.JSON(http.StatusOK, dataRespons)
	} else {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Parameter (balance,type,value) not equal")})
	}
}
func (resource Resource) CashLogs(c echo.Context) error {
	page := c.FormValue("page")
	limit := c.FormValue("limit")
	isAction := c.FormValue("action")
	sortedBy := c.FormValue("sorted_by")
	pageInt, limitInt := ValidatePageAndLimit(page, limit)
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	CashLogsCollection := cashierDatabase.Collection(CashLogsCollectionName)
	sortedByQuery := "created_at"
	query := map[string]interface{}{}
	if sortedBy != "" {
		sortedByQuery = sortedBy
	}
	if isAction != "" {
		query["action"] = isAction
	}

	cashLogDetails, err := GetAllWithPageLimitAndSort(CashLogsCollection, query, pageInt, limitInt, sortedByQuery)
	if err == nil {
		dataRespons := map[string]interface{}{
			"count": len(cashLogDetails),
			"data":  cashLogDetails,
		}
		return c.JSON(http.StatusOK, dataRespons)
	}
	return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Something went wrong error %s", err)})
}

func (resource Resource) Payment(c echo.Context) error {
	cashierID := c.FormValue("cashier_id")
	productIDs := c.FormValue("product_id")
	quantity := c.FormValue("quantity")
	receiveCash := c.FormValue("receive_cash")
	quantityIDArray := strings.Split(quantity, ",")
	productIDArray := strings.Split(productIDs, ",")
	if len(productIDArray) != len(productIDArray) {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Parameter (quantity,product_id) not equal")})
	}
	if receiveCash == "" || receiveCash == "0" {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Parameter (receive_cash) invalid.")})
	}
	if cashierID == "" {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Parameter (cashier_id) missing.")})
	}
	OrderDetail := calculatePrice(productIDArray, quantityIDArray)
	changes, totalChange, err := resource.calculateChange(receiveCash, cashierID, OrderDetail.TotalPrice)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Cannot calculate change because %s", err)})
	}
	OrderDetail.Change = totalChange
	OrderDetail.ReceiveCash = totalChange
	if err := resource.updateCoinStore(changes, cashierID); err != nil {
		return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Cannot update CoinStore because %s", err)})
	}
	CreateCashOrder(OrderDetail)
	dataRespons := make(map[string]interface{})
	dataRespons["total_change"] = totalChange
	dataRespons["change_detail"] = changes
	return c.JSON(http.StatusOK, dataRespons)
}

func (resource Resource) calculateChange(receiveCash, cashierID string, TotalPrice float32) ([]map[string]interface{}, float32, error) {
	var changes []map[string]interface{}
	receiveCashFloat64, err := strconv.ParseFloat(receiveCash, 32)
	if err != nil {
		return changes, 0, err
	}
	receiveCashFloat := float32(receiveCashFloat64)
	if TotalPrice > receiveCashFloat {
		return changes, 0, fmt.Errorf("Not enough cash.")
	}
	totalChange := receiveCashFloat - TotalPrice

	if totalChange > 0 {
		cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
		coinStoreCollection := cashierDatabase.Collection(CoinStoreCollectionName)
		query := primitive.M{
			"cashier_id": cashierID,
			"balance":    primitive.M{"$ne": 0},
		}
		coinStores, err := GetAllWithSort(coinStoreCollection, query, "value") //cashierDetail
		if err != nil {
			return changes, 0, err
		}
		coinStoresJson, _ := json.Marshal(coinStores)
		fmt.Println("coinStoresJson:", string(coinStoresJson))
		fmt.Println("totalChange:", totalChange)
		for _, coinStore := range coinStores {
			if valueInterface, ok := coinStore["value"]; ok {
				value := convertToFloat32(valueInterface)
				change := make(map[string]interface{})
				// if value, ok := coinStore["value"].(int32); ok {
				fmt.Println("=====================================")
				fmt.Println("cash round value:", value)
				valuefloat32 := float32(value)
				changeBalance := totalChange / valuefloat32
				if balanceStore, ok := coinStore["balance"].(int32); ok {
					if int(math.Trunc(float64(changeBalance))) > int(balanceStore) {
						changeBalance = float32(balanceStore)
					}
				}
				if changeBalance >= 1 {
					if totalChange < float32(value) {
						fmt.Println("continue!!!")
						continue
					}
					change["value"] = value
					change["quantity"] = int32(math.Trunc(float64(changeBalance)))
					// fmt.Println("change:", change)
					// fmt.Println("changeBalance:", float32(math.Trunc(float64(changeBalance))))
					// fmt.Println("Balance:", float32(math.Round(float64(valuefloat32))))
					// fmt.Println("changeBalanceOld:", changeBalance)
					// fmt.Println("changeBalanceSum:", (float32(math.Trunc(float64(changeBalance))) * float32(valuefloat32)))
					totalChange = totalChange - (float32(math.Trunc(float64(changeBalance))) * float32(valuefloat32))
					changes = append(changes, change)
				}
			}
			fmt.Println("totalChange:", totalChange)
			if totalChange <= 0 {
				fmt.Println("break!!!")
				break
			}
		}
		if totalChange != 0 {
			return changes, 0, fmt.Errorf("The change is not enough.")
		}
	}
	return changes, (receiveCashFloat - TotalPrice), err
}
func calculatePrice(productIDs, quantity []string) OrderDetail {
	var ProductDetails ProductDetails
	var ProductDetail ProductDetail
	var OrderDetail OrderDetail
	for index, productID := range productIDs {
		if ProductDetailMap, ok := ProductDetials[productID]; ok {
			if id, ok := ProductDetailMap["id"].(int32); ok {
				ProductDetail.ID = id
			}
			if title, ok := ProductDetailMap["title"].(string); ok {
				ProductDetail.Title = title
			}
			if price, ok := ProductDetailMap["price"]; ok {
				ProductDetail.Price = convertToFloat32(price)
			}
			quantity, err := strconv.Atoi(quantity[index])
			if err != nil {

			}
			ProductDetail.TotalPrice = ProductDetail.Price * float32(quantity)
			OrderDetail.TotalPrice += ProductDetail.TotalPrice
			ProductDetails.ProductDetails = append(ProductDetails.ProductDetails, ProductDetail)
		}
	}
	OrderDetail.ProductDetails = ProductDetails
	ProductDetails.Count = len(ProductDetails.ProductDetails)
	return OrderDetail
}
func (resource Resource) updateCoinStore(changes []map[string]interface{}, cashierID string) error {
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	coinStoreCollection := cashierDatabase.Collection(CoinStoreCollectionName)
	query := primitive.M{"cashier_id": cashierID}
	cashDetails, err := resource.Mongo.GetAll(coinStoreCollection, query)
	if err != nil {
		return err
	}
	cashDetailCacher := make(map[float32]map[string]interface{})
	for _, cashDetail := range cashDetails {
		if value, ok := cashDetail["value"]; ok {
			cashDetailCacher[convertToFloat32(value)] = cashDetail
		}
	}
	for _, change := range changes {
		if valueFloat32, ok := change["value"].(float32); ok {
			cashDetail := cashDetailCacher[valueFloat32]
			cashDetailOld := Cloner(cashDetail)
			balanceOld := cashDetail["balance"].(int32)
			fmt.Println("quantity:", change["quantity"])
			fmt.Println(reflect.TypeOf(change["quantity"]))
			if valueInt32, ok := change["quantity"].(int32); ok {
				fmt.Println("quantity:", valueInt32)
				fmt.Println(reflect.TypeOf(valueInt32))
				cashDetail["balance"] = balanceOld - valueInt32
			}
			fmt.Println("balance::", cashDetail["balance"])
			delete(cashDetail, "_id")
			query["value"] = change["value"]
			if _, err := resource.Mongo.UpdateOne(coinStoreCollection, query, primitive.M{"$set": cashDetail}); err != nil {
				return err
			}
			delete(cashDetailOld, "_id")
			auditedChanges := []map[string]interface{}{
				cashDetailOld,
				cashDetail,
			}
			CreateCashLog(actionOrderPayment, auditedChanges)
		}
	}
	return err
}

func (resource Resource) GetOrder(c echo.Context) error {
	page := c.FormValue("page")
	limit := c.FormValue("limit")
	pageInt, limitInt := ValidatePageAndLimit(page, limit)
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	orderLogCollection := cashierDatabase.Collection(OrderLogsCollectionName)
	sortedByQuery := "created_at"
	query := map[string]interface{}{}

	orders, err := GetAllWithPageLimitAndSort(orderLogCollection, query, pageInt, limitInt, sortedByQuery)
	if err == nil {
		dataRespons := map[string]interface{}{
			"count": len(orders),
			"data":  orders,
		}
		return c.JSON(http.StatusOK, dataRespons)
	}
	return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Something went wrong error %s", err)})
}
func (resource Resource) GetLogcash(c echo.Context) error {
	page := c.FormValue("page")
	limit := c.FormValue("limit")
	pageInt, limitInt := ValidatePageAndLimit(page, limit)
	cashierDatabase := resource.Mongo.ChangeSchema(CashierDatabaseName)
	cashLogCollection := cashierDatabase.Collection(CashLogsCollectionName)
	sortedByQuery := "created_at"
	query := map[string]interface{}{}

	cashLogs, err := GetAllWithPageLimitAndSort(cashLogCollection, query, pageInt, limitInt, sortedByQuery)
	if err == nil {
		dataRespons := map[string]interface{}{
			"count": len(cashLogs),
			"data":  cashLogs,
		}
		return c.JSON(http.StatusOK, dataRespons)
	}
	return c.JSON(http.StatusInternalServerError, bson.M{"message": fmt.Sprintf("Something went wrong error %s", err)})
}
