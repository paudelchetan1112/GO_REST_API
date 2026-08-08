package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"restapi/internal/models"
	"restapi/internal/repositories/sqlconnect"
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
}
func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}
func isValidSortField(field string) bool {
	validField := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	}
	return validField[field]

}
func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)

	}
	defer db.Close()

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idstr := strings.TrimSuffix(path, "/")
	if idstr == "" {

		query := "SELECT id, first_name, last_name, email, class, subject From teachers WHERE 1=1"
		var args []interface{}

		query, args = addFilter(r, query, args)
		query = addSorting(r, query)

		rows, err := db.Query(query, args...)
		if err != nil {
			fmt.Println("Error", err)
			http.Error(w, "Database Query error", http.StatusInternalServerError)
			return

		}

		defer rows.Close()

		teachersList := make([]models.Teacher, 0)
		for rows.Next() {
			var teacher models.Teacher
			err := rows.Scan(&teacher.Id, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
			if err != nil {
				http.Error(w, "Error scanning database results", http.StatusInternalServerError)
				return

			}

			teachersList = append(teachersList, teacher)
		}

		response := struct {
			Status string           `json:"Status"`
			Count  int              `json:"count"`
			Data   []models.Teacher `json:"data"`
		}{
			Status: "Success",
			Count:  len(teachersList),
			Data:   teachersList,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		//handle path parameter
		id, err := strconv.Atoi(idstr)
		if err != nil {
			fmt.Println(err)
			return
		}
		var teacher models.Teacher
		err = db.QueryRow("SELECT id, first_name, last_name,email, class, subject FROM teachers WHERE id=?", id).Scan(&teacher.Id, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err == sql.ErrNoRows {

			http.Error(w, "Teacher not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Database query error", http.StatusInternalServerError)

		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(teacher)
	}

}

func addSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		query += "ORDER BY"
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidSortField(field) || !isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				query += ","

			}
			query += " " + field + " " + order
		}
	}
	return query
}

func addFilter(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"class":      "class",
		"subject":    "subject",
	}
	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + dbField + "= ?  "
			args = append(args, value)
		}

	}
	return query, args
}

func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDb()

	if err != nil {
		http.Error(w, "Error connecting to the database", http.StatusInternalServerError)
		return

	}
	defer db.Close()
	stmt, err := db.Prepare("INSERT INTO teachers(first_name, last_name, email, class, subject) VALUES(?, ?, ?, ?,?)")
	if err != nil {
		http.Error(w, "Error in preparing SQL query", http.StatusInternalServerError)

	}

	defer stmt.Close()

	var newTeachers []models.Teacher

	err = json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invlid Request Body", http.StatusBadRequest)
		return

	}

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		if err != nil {
			http.Error(w, "Error inserting data into database", http.StatusInternalServerError)
			return

		}
		lastID, err := res.LastInsertId()
		if err != nil {
			http.Error(w, "Error getting last insert Id", http.StatusInternalServerError)

		}
		newTeacher.Id = int(lastID)
		addedTeachers[i] = newTeacher
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
func updateTeacherHandler(w http.ResponseWriter, r *http.Request) {

	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr = strings.TrimSuffix(idStr, "/")
	fmt.Println(idStr)
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
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		log.Println(err)
		http.Error(w, "unable to connect to database id", http.StatusInternalServerError)

		return

	}
	defer db.Close()
	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id=?", id).Scan(&existingTeacher.Id, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Teacher not found", http.StatusNotFound)
			return

		}

		http.Error(w, "Unable to retrive data", http.StatusInternalServerError)

		return
	}
	updatedTeacher.Id = existingTeacher.Id
	_, err = db.Exec("UPDATE teachers set first_name=?, last_name=?, email=?, class=?, subject=? WHERE id=?", updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Class, updatedTeacher.Subject, updatedTeacher.Id)

	if err != nil {

		http.Error(w, "Teacher not found", http.StatusNotFound)
		return

	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeacher)

}

//PATCH /teachers/id

func patchTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr = strings.TrimSuffix(idStr, "/")
	fmt.Println(idStr)
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
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		log.Println(err)
		http.Error(w, " Unable to connect to database", http.StatusInternalServerError)
		return

	}
	defer db.Close()
	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id=?", id).Scan(&existingTeacher.Id, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err != nil {
		if err == sql.ErrNoRows {

			http.Error(w, "Teacher not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return
	}
	//apply update

	// for k, v := range updates {
	// 	switch k {
	// 	case "first_name":
	// 		existingTeacher.FirstName = v.(string)
	// 	case "last_name":
	// 		existingTeacher.LastName = v.(string)
	// 	case "email":
	// 		existingTeacher.Email = v.(string)
	// 	case "class":
	// 		existingTeacher.Class = v.(string)
	// 	case "subject":
	// 		existingTeacher.Subject = v.(string)
	// 	}
	// }

	//apply update using reflect
	teacherVal := reflect.ValueOf(&existingTeacher).Elem()
	teacherType := teacherVal.Type()

	for k, v := range updates {
		for i := 0; i < teacherVal.NumField(); i++ {
			field := teacherType.Field(i)

			tag := field.Tag.Get("json")
			tagName := strings.Split(tag, ",")[0]
			if tagName == k {
				if teacherVal.Field(i).CanSet() {
					val := reflect.ValueOf(v)
					if val.Type().ConvertibleTo(teacherVal.Field(i).Type()) {
						teacherVal.Field(i).Set(val.Convert(teacherVal.Field(i).Type()))

					}
				}
			}

		}
	}
	_, err = db.Exec("UPDATE teachers set first_name=?, last_name=?, email=?, class=?, subject=? WHERE id=?", existingTeacher.FirstName, existingTeacher.LastName, existingTeacher.Email, existingTeacher.Class, existingTeacher.Subject, existingTeacher.Id)
	if err != nil {

		http.Error(w, "Teacher not found", http.StatusNotFound)
		return

	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingTeacher)

}

func deleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr = strings.TrimSuffix(idStr, "/")
	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "doesn't find Id", http.StatusBadRequest)
		return
	}
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to connect to the database3 ", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	result, err := db.Exec("DELETE FROM teachers WHERE id=?", id)
	if err != nil {

		http.Error(w, "Teacher not found", http.StatusNotFound)
		return

	}
	fmt.Println(result.RowsAffected())
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error while retrieving delete result", http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Teacher not Found", http.StatusNotFound)
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

func TeacherHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		//call get method handler function
		GetTeachersHandler(w, r)

	case http.MethodPost:

		//call post method handler function
		AddTeacherHandler(w, r)

	case http.MethodDelete:
		deleteTeacherHandler(w, r)

	case http.MethodPut:
		//call post put method handler method
		updateTeacherHandler(w, r)

	case http.MethodPatch:

		patchTeacherHandler(w, r)

	default:
		w.Write([]byte("Invalid request method"))

	}

}
