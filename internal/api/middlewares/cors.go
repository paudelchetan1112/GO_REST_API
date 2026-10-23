package middlewares

import (
	"fmt"
	"net/http"
)
var allowedOrigins =[]string{
	"https://localhost:3000",
	"https://myorigin.com",
	
}
func Cors(next http.Handler) http.Handler{
		fmt.Println("cors middleware.....")
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("cors middleware return.....")
origin:=r.Header.Get("Origin")
fmt.Println(origin)
if isallowedOrigin(origin){
w.Header().Set("Access-Control-Allow-Origin", origin)
}else{
	http.Error(w, "not allowed",http.StatusForbidden)

	return
}
w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
w.Header().Set("Access-Control-Allow-Credentials", "true")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
w.Header().Set("Access-Control-Expose-Headers", "Authorization")
w.Header().Set("Access-Control-Max-Age", "3600")
	if r.Method==http.MethodOptions{

	}


	next.ServeHTTP(w,r)
		fmt.Println("cors middleware after end.....")
})
}

func isallowedOrigin(origin string)bool{
	for _, allowedOrigin:=range allowedOrigins{
		if origin==allowedOrigin{
			return true
		}
	}
	return false
}

// package middlewares

// import (
// 	"fmt"
// 	"net/http"
// )

// var allowedOrigins=[]string{
// 	"https://my-origin-url.com",
// 	"https://localhost:3000",

// }
// func Cors(next http.Handler)http.Handler{
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		origin:=r.Header.Get("Origin")
// 		fmt.Println(origin)
// 		if isOriginAllowed(origin){
// w.Header().Set("Access-Control-Allow-Origin", origin)

// 		}else{
// 			http.Error(w, "Not allowed by CORS", http.StatusForbidden)
// 		}

// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
// 		w.Header().Set("Access-Control-Allow-Credentials", "true")
// 		w.Header().Set("Access-Control-Allow-Credential", "true")
// 		w.Header().Set("Access-Control-Max-Age", "3600")

// 		if r.Method==http.MethodOptions{

// 		}

// 		next.ServeHTTP(w,r)

// 	})
// }

// func isOriginAllowed(origin string)bool{
// 	for _, allowedOrigin:=range allowedOrigins{
// 		if origin==allowedOrigin{
// 			return true
// 		}
// 	}
// 	return false

// }