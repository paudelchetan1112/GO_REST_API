package middlewares

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]int
	limit     int
	resetTime time.Duration
}

func NewRateLimiter(limit int, resetTime time.Duration) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]int),
		limit:    limit,
		resetTime: resetTime,

	}
	go rl.resetVisitorCount()

	return rl

}
func (rl *rateLimiter) resetVisitorCount(){
	for{
		time.Sleep(rl.resetTime)
		rl.mu.Lock()
		rl.visitors=make(map[string]int)
		rl.mu.Unlock()

	}
}


func (rl *rateLimiter) Middleware (next http.Handler) http.Handler{
	fmt.Println("Rate limiting middleware.....")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Rate limiting middleware return.....")

rl.mu.Lock()
defer rl.mu.Unlock()

visitorIp:=r.RemoteAddr //you might want to extract the Ip in a more sophisticated way
rl.visitors[visitorIp]++
fmt.Printf("Visitor count from %v is %v\n", visitorIp, rl.visitors[visitorIp])

if rl.visitors[visitorIp]>rl.limit{
	http.Error(w, "Too many requests", http.StatusTooManyRequests)
	return 

}

next.ServeHTTP(w,r)
fmt.Println("Rate limiting middleware after next")

	})
}