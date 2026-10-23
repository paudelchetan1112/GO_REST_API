package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"restapi/internal/models"
	"restapi/internal/repositories/sqlconnect"
	"strconv"
	"strings"
)

func GetStudentHandler(w http.ResponseWriter, r *http.Request) {
	var students []models.Student
	students, err := sqlconnect.GetStudentsDbHandler(students, r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}
	reponse := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "Success",
		Count:  len(students),
		Data:   students,
	}
	w.Header().Set("content-Type", "application/json")
	json.NewEncoder(w).Encode(reponse)

}

func AllowedJSONFields(model any) (map[string]struct{}, error) {
	t := reflect.TypeOf(model)

	if t == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a struct")
	}

	allowed := make(map[string]struct{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// JSON does not expose unexported fields.
		if field.PkgPath != "" {
			continue
		}

		tag := field.Tag.Get("json")
		parts := strings.Split(tag, ",")
		name := parts[0]

		if name == "-" {
			continue
		}

		// Match encoding/json's default field-name behavior.
		if name == "" {
			name = field.Name
		}

		allowed[name] = struct{}{}
	}

	return allowed, nil
}

func ValidateRequired(model any) error {
	if model == nil {
		return fmt.Errorf("model cannot be nil")
	}

	v := reflect.ValueOf(model)

	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fmt.Errorf("model cannot be nil")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct")
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)

		if fieldType.Tag.Get("validate") != "required" {
			continue
		}

		if fieldValue.IsZero() {
			return fmt.Errorf("%s is required", fieldType.Name)
		}
	}

	return nil
}

func AddStudentHandler(w http.ResponseWriter, r *http.Request) {
	var newStudents []models.Student
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&newStudents)
	if err != nil {
		http.Error(w, "invalid Request Body", http.StatusBadRequest)
		return
	}

	// allowedFields, err := AllowedJSONFields(models.Teacher{})
	// if err != nil {
	//     http.Error(w, "Invalid teacher model", http.StatusInternalServerError)
	//     return
	// }

	for _, student := range newStudents {
		if err := ValidateRequired(student); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	addedStudents := make([]models.Student, len(newStudents))
	addedStudents, err = sqlconnect.AddStudentDbHandler(newStudents)

	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string           `json:"status"`
		Count  int              `json:"count"`
		Data   []models.Student `json:"data"`
	}{
		Status: "success",
		Count:  len(addedStudents),
		Data:   addedStudents,
	}
	json.NewEncoder(w).Encode(response)

}

func GetOneStudent(w http.ResponseWriter, r *http.Request) {
	idstr := r.PathValue("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	student, err := sqlconnect.GetOneStudentDbHandler(id)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := struct {
		Status string         `json:"status"`
		Data   models.Student `json:"data"`
	}{
		Status: "success",
		Data:   student,
	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func DeleteOneStudent(w http.ResponseWriter, r *http.Request) {
	idstr := r.PathValue("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = sqlconnect.DeleteOneStudentDbHandler(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := struct {
		Status           string `json:"status"`
		DeletedStudentId int    `json:"deletedstudentid"`
	}{
		Status:           "Deleted successfully",
		DeletedStudentId: id,
	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(response)

}
func DeleteMultipleStudent(w http.ResponseWriter, r *http.Request) {
	var StudentIDs []int
	err := json.NewDecoder(r.Body).Decode(&StudentIDs)
	if err != nil {
		log.Println(err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	deletedIds, err := sqlconnect.DeleteStudentDbHandler(StudentIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := struct {
		Status     string `json:"status"`
		DeletedIds []int  `json:"deleted_ids"`
	}{
		Status:     "Teacher successfully deleted",
		DeletedIds: deletedIds,
	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(response)

}
func UpdateStudentHandler(w http.ResponseWriter, r *http.Request) {
	idstr := r.PathValue("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		log.Println("invalid Input")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload models.Student
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&payload)
	if err != nil {
		log.Println("Error while decoding")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// (Checks if the payload ID was sent [not 0] and doesn't match the URL)
	if payload.Id != 0 && payload.Id != id {
		log.Println("Data integrity error: URL ID and JSON Payload ID mismatch")
		http.Error(w, "URL ID and JSON Payload ID mismatch", http.StatusBadRequest)
		return
	}
	updatedStudent, err := sqlconnect.UpdateStudentDbHandler(id, payload)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := struct {
		Status string         `json:"status"`
		Data   models.Student `json:"data"`
	}{
		Status: "Update Successful",
		Data:   updatedStudent,
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}
func PatchOneStudentHandler(w http.ResponseWriter, r *http.Request) {
	idstr := r.PathValue("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		log.Println("invalid Input")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Println("Error while decoding")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	delete(payload, "id")
	if len(payload) == 0 {
		http.Error(w, "No valid fields provided for patch update", http.StatusBadRequest)
		return
	}
	updatedStudent, err := sqlconnect.PatchOneStudentDbHandler(id, payload)
	responese := struct {
		Status string         `json:"string"`
		Data   models.Student `json:"data"`
	}{
		Status: "Update Successful",
		Data:   updatedStudent,
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responese)

}

func PatchStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var payload []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Println("Error while decoding")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updatedStudent, err := sqlconnect.PatchStudentDbHandler(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status         string           `json:"status"`
		Count          int              `json:"count"`
		UpdatedStudent []models.Student `json:"updatedstudent"`
	}{
		Status:         "Successfully updated",
		Count:          len(updatedStudent),
		UpdatedStudent: updatedStudent,
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

// import "net/http"

// func StudentHandler(w http.ResponseWriter, r *http.Request) {

// 	switch r.Method {
// 	case http.MethodGet:
// 		w.Write([]byte("Hello Get method on student Route"))

// 	case http.MethodPost:
// 		w.Write([]byte("Hello POST method on student Route"))

// 	case http.MethodDelete:
// 		w.Write([]byte("Hello Delete  method on student Route"))

// 	case http.MethodPut:
// 		w.Write([]byte("Hello Put method on student Route"))

// 	case http.MethodPatch:
// 		w.Write([]byte("Hello Patch Method on student Route"))

// 	default:
// 		w.Write([]byte("Invalid request method"))

// 	}
// }
// func GetStudentsHandler(w http.ResponseWriter, r *http.request){
// 	idStr:=r.PathValue("id")
// 	id, err:=stringconv.Atoi(idstr)
// 	if err!=nil{
// 	http.Error("Internal Server error", http.StatusInternalServerError)
// return
// }
// 	var student models.Student

// }
