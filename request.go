//Access the request detail:
fmt.Println("Body:", r.Body)
fmt.Println("Form:", r.Form)
fmt.Println("Header", r.Header)
fmt.Println("Context:", r.Context())
fmt.Println("Contentlength:",r.ContentLength)
fmt.Println("Host:", r.Host)
fmt.Println("method:", r.Method)
fmt.Println("Protocol:", r.Proto)
fmt.Println("Remote URI:", r.RequestURI)
fmt.Println("transfer Encoding:", r.TransferEncoding)

fmt.Println("URL", r.URL)
fmt.Println("User Agent:", r.UserAgent())

fmt.Println("Path", r.URL.Path)
fmt.Println("Params", r.URL.Port())
fmt.Println("Querys", r.URL.Query())
fmt.Println("URL Scheme", r.URL.Scheme)