package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	mw "restapi/internal/api/middlewares"
	"restapi/internal/api/router"
	"restapi/internal/repositories/sqlconnect"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment configurations
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// 2. Initialize database connection
	_, err = sqlconnect.ConnectDb()
	if err != nil {
		log.Fatalf("Error while connecting db: %v", err)
	}

	// 3. Define server configuration details
	port := ":3000"
	cert := "certificate.crt"
	key := "server.key"

	// 4. Configure high-grade TLS standards
	tlsConfig := &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_CHACHA20_POLY1305_SHA256,
		},
	}

	// 5. Initialize Middleware Configurations
	rl := mw.NewRateLimiter(5, time.Minute)
	hppOption := mw.HPPOptions{
		CheckQuery:                  true,
		CheckBody:                   true,
		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
		Whitelist:                   []string{"allowedParam", "sortBy", "sortOrder", "name", "age", "class"},
	}

	// 6. Initialize Routes & Conditional Middlewares
	r := router.MainRouter()
	jwtMiddleware := mw.MiddlewaresExcludePaths(mw.JWTMiddleware, "/execs/login")

	// 7. Assemble Production Middleware Chain (Outside-In Execution Sequence)
	// Order of execution: CORS -> Rate Limiter -> Response Time -> Security Headers -> Compression -> HPP -> JWT -> Router
	secureMux := mw.Cors(
		rl.Middleware(
			mw.ResponseTimeMiddleware(
				mw.SecurityHeaders(
					mw.Compression(
						mw.Hpp(hppOption)(
							jwtMiddleware(r),
						),
					),
				),
			),
		),
	)

	// 8. Create Custom Production Server Instance
	server := &http.Server{
		Addr:              port,
		Handler:           secureMux,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Println("Secure production server is running on port:", port)
	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting the server:", err)
	}
}

// package main

// import (
// 	"crypto/tls"
// 	"fmt"

// 	// "sync"

// 	// "time"

// 	// "io"
// 	"log"
// 	"net/http"

// 	// "strings"


// 	sqlconnect "restapi/internal/repositories/sqlconnect"

// 	"github.com/joho/godotenv"
// )

// func main() {
// 	err := godotenv.Load()
// 	if err != nil {

// 		return
// 	}
// 	_, err = sqlconnect.ConnectDb()

// 	if err != nil {
// 		fmt.Println("Error while connecting db")
// 		return
// 	}

// 	port := ":3000"
// 	cert := "certificate.crt"
// 	key := "server.key"

// 	tlsConfig := &tls.Config{
// 		MinVersion: tls.VersionTLS12,
// 	}
// 	// rl:=mw.NewRateLimiter(5, time.Minute)

// 	// hppOption:=mw.HPPOptions{
// 	// 	CheckQuery: true,
// 	// 	CheckBody: true,
// 	// 	CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
// 	// 	Whitelist: []string{"allowedParam", "sortBy", "sortOrder","name", "age", "class"},

// 	// }
// 	// secureMux:= mw.Cors(rl.Middleware(mw.ResponseTimeMiddleware(mw.SecurityHeaders(mw.Compression(mw.Hpp(hppOption)(mux))))))

// 	// secureMux:=applyMiddlewares(mux, mw.Hpp(hppOption), mw.Compression, mw.SecurityHeaders, mw.ResponseTimeMiddleware, rl.Middleware, mw.Cors)
// 	router := router.MainRouter()
// 	jwtMiddleware := mw.MiddlewaresExcludePaths(mw.JWTMiddleware, "/execs/login")
// 	// secureMux := utils.ApplyMiddlewares(router, mw.SecurityHeaders)

// 	secureMux := jwtMiddleware(mw.SecurityHeaders(router))
// 	//create custom server
// 	server := &http.Server{
// 		Addr:    port,
// 		Handler: secureMux,
// 		// Handler: middlewares.Cors(mux),

// 		TLSConfig: tlsConfig,
// 	}
// 	fmt.Println("Server is running on port:", port)

// 	err = server.ListenAndServeTLS(cert, key)
// 	if err != nil {
// 		log.Fatalln("Error starting the server", err)

// 	}

// }

// //Middleware is a function that wraps an http.Handler with additional functionality
