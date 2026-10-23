package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

func ResponseTimeMiddleware(next http.Handler) http.Handler {
	fmt.Println("Response Time Middleware.... ")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Response Time Middleware return.... ")

		start := time.Now()

		//create a ustom responsewriter to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)
		duration := time.Since(start)
		w.Header().Set("X-Response-Time", duration.String())
		
		//log the request detaio

		fmt.Printf("Method:%s, URL:%s, Status:%d, Duration:%v\n", r.Method, r.URL, wrappedWriter.status, duration.String())
		fmt.Println("Sent Response from Response Time Middleware after next")

	})
}

// responseWriter
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
