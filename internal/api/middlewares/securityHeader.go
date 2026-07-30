package middlewares

import (
	"fmt"
	"net/http"
)

func SecurityHeaders(next http.Handler) http.Handler {
	fmt.Println("Security Headers middleware....")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Security Headers middleware return....")

		w.Header().Set("X-DNS-Prefetch-Control", "off")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referred-Policy", "no-referrer")

		w.Header().Set("x-powered-By", "express")

		w.Header().Set("Server", "")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
	
		w.Header().Set("Access-Control-Allow-Credential", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Permissions-Policy", "geolocation=(self), microphone=()")

		next.ServeHTTP(w, r)
	fmt.Println("Security Headers middleware after next ....")


	})

}

//BASIC MIDDLEWARE SKELETON

// func securityHeaders(next http.Handler)http.Handler{
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// next.ServeHTTP(w,r)
// })
// }

// func SecureHeaders(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set()

// 		next.ServeHTTP(w, r)

// 	})
// }
