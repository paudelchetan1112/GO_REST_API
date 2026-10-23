package router

import (
	"net/http"
	"restapi/internal/api/handlers"
)
func studentRouter()*http.ServeMux{
	mux:=http.NewServeMux()
	
	mux.HandleFunc("GET /students", handlers.GetStudentHandler)

	mux.HandleFunc("POST /students", handlers.AddStudentHandler)
	mux.HandleFunc("PATCH /students", handlers.PatchStudentsHandler)
	mux.HandleFunc("DELETE /students", handlers.DeleteMultipleStudent)
	mux.HandleFunc("PUT /students/{id}", handlers.UpdateStudentHandler)
	mux.HandleFunc("PATCH /students/{id}", handlers.PatchOneStudentHandler)

	mux.HandleFunc("DELETE /students/{id}", handlers.DeleteOneStudent)
mux.HandleFunc("GET /students/{id}", handlers.GetOneStudent)
	


	return mux
}