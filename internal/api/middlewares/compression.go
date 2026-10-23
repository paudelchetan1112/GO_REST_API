package middlewares

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
)

func Compression(next http.Handler)http.Handler{
	fmt.Println("compression middleware.....")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//check of the client accepts gzip encoding
fmt.Println("compression middleware return.....")
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip"){
			next.ServeHTTP(w,r)
		}
//Set the response header
w.Header().Set("Content-Encoding", "gzip")
gz:=gzip.NewWriter(w)
defer gz.Close()

//wrap the Response Writer

w=&gzipResponseWriter{ResponseWriter: w, Writer:gz }

		next.ServeHTTP(w,r)
		fmt.Println("Sent response from compression middleware after next")

	})
}

//gzipResponseWriter wraps http.ResponseWriter to write gzipped response

type gzipResponseWriter struct{
	http.ResponseWriter
	Writer *gzip.Writer

}
func (g *gzipResponseWriter) Write(b []byte)(int, error){
return g.Writer.Write(b)

}