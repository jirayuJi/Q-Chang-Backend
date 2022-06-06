package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"golang.org/x/net/context"
)

// Env is global variable for set environment
var Env string

// Port is global variable for start server
var Port string

var CashierID string

var Mode string

var configuration Configuration

var resource Resource

func init() {
	flag.StringVar(&Port, "port", "8081", "The address to listen on for HTTP requests")
	flag.Parse()
	configuration = GetConfiguration()
}

func main() {
	configuration.Mongo.Session = configuration.Mongo.Connect("primary")
	defer configuration.Mongo.Session.EndSession(context.Background())

	resource = Resource{
		Mongo: configuration.Mongo,
	}
	resource.cacherProduct()
	go func() {
		for range time.Tick(time.Minute * 5) {
			resource.cacherProduct()
		}
	}()

	Echo := echo.New()
	resource.initialRouting(Echo)
	s := &http.Server{
		Addr:         fmt.Sprintf(":%s", Port),
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  300 * time.Second,
	}
	Echo.StartServer(s)
}
