package router

import (
	"net/http"
	// "restapi/internal/api/handlers"
)

func MainRouter() *http.ServeMux {
	eRouter:=execsRouter()
	tRouter:=teacherRouter()
	sRouter:=studentRouter()
	sRouter.Handle("/", eRouter)
	tRouter.Handle("/", sRouter) 
	return tRouter

	
}
