package middlewares

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"restapi/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
)

type ContextKey string

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Read the JWT cookie
		cookie, err := r.Cookie("Bearer")
		if err != nil {
			http.Error(w, "Authorization cookie missing", http.StatusUnauthorized)
			return
		}

		// 2. Read the secret key
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			log.Println("JWT_SECRET is not configured")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 3. Parse and validate the JWT
		parsedToken, err := jwt.Parse(
			cookie.Value,
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf(
						"unexpected signing method: %v",
						token.Header["alg"],
					)
				}
				return []byte(jwtSecret), nil
			},
			jwt.WithValidMethods([]string{
				jwt.SigningMethodHS256.Alg(),
			}),
		)

		// 4. Handle parsing and validation errors
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}

			if errors.Is(err, jwt.ErrTokenMalformed) {
				http.Error(w, "Token malformed", http.StatusUnauthorized)
				return
			}

			utils.ErrorHandler(err, "")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// 5. Confirm the token is valid
		if !parsedToken.Valid {
			http.Error(w, "Invalid login token", http.StatusUnauthorized)
			return
		}

		// 6. Extract claims
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		log.Println("Valid JWT")

		// 7. Add claims to the request context
		ctx := context.WithValue(
			r.Context(),
			ContextKey("role"),
			claims["role"],
		)

		ctx = context.WithValue(
			ctx,
			ContextKey("expiresAt"),
			claims["exp"],
		)

		ctx = context.WithValue(
			ctx,
			ContextKey("username"),
			claims["user"],
		)

		ctx = context.WithValue(
			ctx,
			ContextKey("userId"),
			claims["uid"],
		)

		// 8. Pass the updated request to the next handler
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}