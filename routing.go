package main

import (
	"github.com/labstack/echo"
)

func (resource Resource) initialRouting(e *echo.Echo) {
	v1 := e.Group("/v1")
	v1.GET("/", resource.healthcheck)
	v1.GET("/healthcheck", resource.healthcheck)
	v1.GET("/cashiers", resource.GetCashiers)
	v1.POST("/cashier", resource.CreateCashier)
	v1.PUT("/cashier", resource.UpdateCashier)
	v1.POST("/top_up", resource.TopUp)
	v1.GET("/cash_log", resource.CashLogs)
	v1.POST("/payment", resource.Payment)
	v1.GET("/order_log", resource.GetOrder)
	v1.GET("/cash_log", resource.GetLogcash)
}
