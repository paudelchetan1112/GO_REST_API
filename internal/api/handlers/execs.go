package handlers

import "net/http"

func ExecsHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello Get method on Execs Route"))

	case http.MethodPost:
		w.Write([]byte("Hello post method on Execs Route"))

	case http.MethodDelete:
		w.Write([]byte("Hello Delete  method on Execs Route"))

	case http.MethodPut:
		w.Write([]byte("Hello Put method on Execs Route"))

	case http.MethodPatch:
		w.Write([]byte("Hello Patch Method on Execs Route"))

	default:
		w.Write([]byte("Invalid request method"))

	}
}