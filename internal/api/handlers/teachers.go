package handlers

import (
	// "database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	// "reflect"

	"restapi/internal/models"
	"restapi/internal/repositories/sqlconnect"
	"strconv"

	// "strings"
	"sync"
)

var (
	teachers = make(map[int]models.Teacher)
	mutex    = &sync.Mutex{}
	nextId   = 1
)

func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var teachers []models.Teacher

	teachers, err := sqlconnect.GetTeachersDbHandler(teachers, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string           `json:"Status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "Success",
		Count:  len(teachers),
		Data:   teachers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func GetOneTeacherHandler(w http.ResponseWriter, r *http.Request) {

	idstr := r.PathValue("id")

	//handle path parameter
	id, err := strconv.Atoi(idstr)
	if err != nil {
		fmt.Println(err)
		return
	}
	teacher, err := sqlconnect.GetOneTeacherDbHandler(w, id)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(teacher)
}

func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {

	var newTeachers []models.Teacher

	err := json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invlid Request Body", http.StatusBadRequest)
		return

	}
	val:=reflect.TypeOf(models.Teacher{})
	fields:=[]string{}
	for i:=0;i<val.NumField();i++{
		field:=val.Field(i)
		fields=append(fields, field.Tag.Get("json"))

	}
	// allowedFields:=make(map[string]struct{})
	// for _, field:=range newTeachers{
	// 	for key:=range teacher{
	// 		_, ok:=allowedFields[key]
	// 		if !ok{
	// 			http.Error(w, "Unacceptable field found in request. Only use allowed Field", http.StatusBadRequest)
	// 		}
	// 	}
	// }


	for _, teacher:=range newTeachers{
		// if teacher.FirstName==""||teacher.LastName==""||teacher.Email==""||teacher.Class==""||teacher.Subject==""{
		// 	http.Error(w, "All field are required", http.StatusBadRequest)
		// 	return 
			
		// }
		val:=reflect.ValueOf(teacher)
		for i:=0;i<val.NumField();i++{
			field:=val.Field(i)
			if field.Kind()==reflect.String &&field.String()==""{
				fmt.Println("Field.kind():", field.Kind())
				fmt.Println("Reflect.string", reflect.String)
				fmt.Println("Field.string()", field.String())
				http.Error(w, "All field are required", http.StatusBadRequest)
				return
			}
		}
	}

	addedTeachers, err := sqlconnect.AddTeacherDbHandler(newTeachers)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(response)

}

// PUT /teachers/
func UpdateTeacherHandler(w http.ResponseWriter, r *http.Request) {

	// idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	// idStr = strings.TrimSuffix(idStr, "/")
	idStr := r.PathValue("id")
	// fmt.Println(idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher Id", http.StatusBadRequest)
		return

	}
	var updatedTeacher models.Teacher
	err = json.NewDecoder(r.Body).Decode(&updatedTeacher)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return

	}
	updatedTeacherFromDB, err := sqlconnect.UpdateTeacherDbHandler(id, updatedTeacher)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacherFromDB)

}

//Patch /teachers

// PATCH /teachers/id
func PatchTeachersHandler(w http.ResponseWriter, r *http.Request) {

	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	err = sqlconnect.PatchTeachersDbHandler(updates)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	w.WriteHeader(http.StatusNoContent)
}

func PatchOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	// idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	// idStr = strings.TrimSuffix(idStr, "/")
	// fmt.Println(idStr)
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher Id", http.StatusBadRequest)
		return

	}
	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return

	}
	updatedTeacher, err := sqlconnect.PatchOneTeacherDbHandler(id, updates)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacher)

}

func DeleteOneTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idstr := r.PathValue("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
	}
	err = sqlconnect.DeleteOneTeacherDbHandler(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	// w.WriteHeader(http.StatusNoContent)
	//Response body
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"Status"`
		Id     int    `json:"id"`
	}{
		Status: "Teacher Successfully deleted",
		Id:     id,
	}
	json.NewEncoder(w).Encode(response)

}

func DeleteTeachersHandler(w http.ResponseWriter, r *http.Request) {

	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload ", http.StatusBadRequest)

		return

	}
	deletedIds, err := sqlconnect.DeleteTeachersDbHandler(ids)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "Application/json")
	response := struct {
		Status     string `json:"status"`
		DeletedIds []int  `json:"deleted_ids"`
	}{
		Status:     "Teacher successfully deleted",
		DeletedIds: deletedIds,
	}

	json.NewEncoder(w).Encode(response)
}
func GetStudentsByTeacherId(w http.ResponseWriter, r *http.Request){
	teacherId:=r.PathValue("id")
	id, err:=strconv.Atoi(teacherId)
if err!=nil{
	log.Println("Invalid teacher ID format:", err)
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

	  var students []models.Student
	students, err=sqlconnect.GetStudentByTeacherId(id)
	if err!=nil{
http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	
response:=struct{
Status string `json:"status"`
Count int `json:"count"`
Data []models.Student `json:"data"`

}{
	Status: "Success",
	Count : len(students),
	Data: students,
}
w.Header().Set("content-type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(response)


	  
}

func GetStudentCountByTeacherId(w http.ResponseWriter, r *http.Request){
		teacherId:=r.PathValue("id")
	id, err:=strconv.Atoi(teacherId)
if err!=nil{
	log.Println("Invalid teacher ID format:", err)
		http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
		return
	}

		  var studentcount int
	studentcount, err=sqlconnect.GetStudentCountByTeacherIdDbHandler(id)
	if err!=nil{
http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	
response:=struct{
Status string `json:"status"`
StudentCount int `json:"studentcount"`

}{
	Status: "Success",
	StudentCount: studentcount,
}
w.Header().Set("content-type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(response)


}

// func TeacherHandler(w http.ResponseWriter, r *http.Request) {

// 	switch r.Method {
// 	case http.MethodGet:
// 		//call get method handler function
// 		GetTeachersHandler(w, r)

// 	case http.MethodPost:

// 		//call post method handler function
// 		AddTeacherHandler(w, r)

// 	case http.MethodDelete:
// 		deleteTeacherHandler(w, r)

// 	case http.MethodPut:
// 		//call post put method handler method
// 		updateTeacherHandler(w, r)

// 	case http.MethodPatch:

// 		patchTeacherHandler(w, r)

// 	default:
// 		w.Write([]byte("Invalid request method"))

// 	}

// }
