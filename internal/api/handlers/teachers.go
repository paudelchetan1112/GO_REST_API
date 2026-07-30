package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"restapi/internal/models"
	"strconv"
	"strings"
	"sync"
)
var (
	teachers = make(map[int]models.Teacher)
	mutex    = &sync.Mutex{}
	nextId   = 1
)

//Initialize some dummy data

func init() {
	teachers[nextId] = models.Teacher{
		Id:        nextId,
		FirstName: "john",
		LastName:  "Doe",
		Class:     "9A",
		Subject:   "Math",
	}
	nextId++
	teachers[nextId] = models.Teacher{
		Id:        nextId,
		FirstName: "Jane",
		LastName:  "Smith",
		Class:     "10A",
		Subject:   "Algebra",
	}

	nextId++
	teachers[nextId] = models.Teacher{
		Id:        nextId,
		FirstName: "Chetan",
		LastName:  "Paudel",
		Class:     "10A",
		Subject:   "Algebra",
	}
	nextId++
}

func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idstr := strings.TrimSuffix(path, "/")
	if idstr == "" {

		firstName := r.URL.Query().Get("first_name")
		lastName := r.URL.Query().Get("last_name")

		teachersList := make([]models.Teacher, 0, len(teachers))

		for _, teacher := range teachers {
			if (firstName == "" || teacher.FirstName == firstName) && (lastName == "" || teacher.LastName == lastName) {
				teachersList = append(teachersList, teacher)
			}

		}

		response := struct {
			Status string    `json:"Status"`
			Count  int       `json:"count"`
			Data   []models.Teacher `json:"data"`
		}{
			Status: "Success",
			Count:  len(teachersList),
			Data:   teachersList,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
	//handle path parameter
	id, err := strconv.Atoi(idstr)
	if err != nil {
		fmt.Println(err)
		return
	}
	teacher, exists := teachers[id]
	if !exists {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return

	}
	json.NewEncoder(w).Encode(teacher)

}

func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	var newTeachers []models.Teacher

	err := json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invlid Request Body", http.StatusBadRequest)
		return

	}

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		newTeacher.Id = nextId
		teachers[nextId] = newTeacher
		addedTeachers[i] = newTeacher
		nextId++

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string    `json:"status"`
		Count  int       `json:"count"`
		Data   []models.Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(response)

}

func TeacherHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		//call get method handler function
		GetTeachersHandler(w, r)

	case http.MethodPost:

		//call post method handler function
		AddTeacherHandler(w, r)

	case http.MethodDelete:
		w.Write([]byte("Hello Delete  method on Teacher Route"))

	case http.MethodPut:
		w.Write([]byte("Hello Put method on Teacher Route"))

	case http.MethodPatch:
		w.Write([]byte("Hello Patch Method on Teacher Route"))

	default:
		w.Write([]byte("Invalid request method"))

	}

}