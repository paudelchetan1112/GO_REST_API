package middlewares

import (
	"fmt"
	"net/http"
	"strings"
)

type HPPOptions struct {
	CheckQuery                  bool
	CheckBody                   bool
	CheckBodyOnlyForContentType string
	Whitelist                   []string
}

func Hpp(options HPPOptions) func(http.Handler) http.Handler {
		fmt.Println("http middleware.....")

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("http middleware return.....")
			// Filter body parameters
			if options.CheckBody &&
				(r.Method == http.MethodPost ||
					r.Method == http.MethodPut ||
					r.Method == http.MethodPatch) &&
				isCorrectContentType(r, options.CheckBodyOnlyForContentType) {

				filterBodyParams(r, options.Whitelist)
			}

			// Filter query parameters
			if options.CheckQuery {
				filterQueryParams(r, options.Whitelist)
			}

			next.ServeHTTP(w, r)
				fmt.Println("http middleware after next.....")
		})
	}
}

func isCorrectContentType(r *http.Request, contentType string) bool {
	if contentType == "" {
		return true
	}
	return strings.Contains(r.Header.Get("Content-Type"), contentType)
}

func filterBodyParams(r *http.Request, whitelist []string) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		return
	}

	for k, v := range r.PostForm {
		// Collapse duplicate values unless whitelisted
		if len(v) > 1 && !isWhiteListed(k, whitelist) {
			r.PostForm.Set(k, v[0])
		}
	}
}

func filterQueryParams(r *http.Request, whitelist []string) {
	query := r.URL.Query()

	for k, v := range query {
		// Collapse duplicate values unless whitelisted
		if len(v) > 1 && !isWhiteListed(k, whitelist) {
			query.Set(k, v[0])
		}
	}

	r.URL.RawQuery = query.Encode()
}

func isWhiteListed(param string, whitelist []string) bool {
	for _, v := range whitelist {
		if param == v {
			return true
		}
	}
	return false
}
