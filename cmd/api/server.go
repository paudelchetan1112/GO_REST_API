package main

import (
	"crypto/tls"
	"fmt"

	// "sync"

	// "time"

	// "io"
	"log"
	"net/http"

	// "strings"

	mw "restapi/internal/api/middlewares"
	"restapi/internal/api/router"
	sqlconnect "restapi/internal/repositories/sqlconnect"

	// "restapi/internal/models"

	"restapi/pkg/utils"

	"github.com/joho/godotenv"
)

func main() {
	err:=godotenv.Load()
	if err!=nil{
	
		return
	}
	_, err=sqlconnect.ConnectDb()

	if err!=nil{
		fmt.Println("Error while connecting db")
		return
	} 

	port := ":3000"
	cert := "certificate.crt"
	key := "server.key"

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	// rl:=mw.NewRateLimiter(5, time.Minute)

	// hppOption:=mw.HPPOptions{
	// 	CheckQuery: true,
	// 	CheckBody: true,
	// 	CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
	// 	Whitelist: []string{"allowedParam", "sortBy", "sortOrder","name", "age", "class"},

	// }
	// secureMux:= mw.Cors(rl.Middleware(mw.ResponseTimeMiddleware(mw.SecurityHeaders(mw.Compression(mw.Hpp(hppOption)(mux))))))

	// secureMux:=applyMiddlewares(mux, mw.Hpp(hppOption), mw.Compression, mw.SecurityHeaders, mw.ResponseTimeMiddleware, rl.Middleware, mw.Cors)
	router := router.Router()
	secureMux := utils.ApplyMiddlewares(router, mw.SecurityHeaders)
	//create custom server
	server := &http.Server{
		Addr:    port,
		Handler: secureMux,
		// Handler: middlewares.Cors(mux),

		TLSConfig: tlsConfig,
	}
	fmt.Println("Server is running on port:", port)

	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting the server", err)

	}

}

//Middleware is a function that wraps an http.Handler with additional functionality
