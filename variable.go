package main

// CoinsStoreInit init
var (
	CoinsStoreInits = []map[string]interface{}{
		{
			"value":   1000,
			"type":    "bank_note",
			"balance": 10,
		},
		{
			"value":   500,
			"type":    "bank_note",
			"balance": 20,
		},
		{
			"value":   100,
			"type":    "bank_note",
			"balance": 15,
		},
		{
			"value":   50,
			"type":    "bank_note",
			"balance": 20,
		},
		{
			"value":   20,
			"type":    "bank_note",
			"balance": 30,
		},
		{
			"value":   10,
			"type":    "coin_value",
			"balance": 20,
		},
		{
			"value":   5,
			"type":    "coin_value",
			"balance": 20,
		},
		{
			"value":   1,
			"type":    "coin_value",
			"balance": 20,
		},
		{
			"value":   0.25,
			"type":    "coin_value",
			"balance": 50,
		},
	}
)

//Global variable
var (
	ProductDetials = make(map[string]map[string]interface{})
)
