package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"

	// "log"
	"net/http"
	"reflect"
	"restapi/internal/models"
	"restapi/pkg/utils"
	"strconv"
	"strings"
)

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
func GetTeachersDbHandler(teachers []models.Teacher, r *http.Request) ([]models.Teacher, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")

	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class, subject From teachers WHERE 1=1"
	var args []interface{}

	query, args = addFilter(r, query, args)
	query = addSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println("Error", err)
		return nil, utils.ErrorHandler(err, "Database Query error")

	}

	defer rows.Close()

	// teachersList := make([]models.Teacher, 0)
	for rows.Next() {
		var teacher models.Teacher
		err := rows.Scan(&teacher.Id, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error Scanning database results")

		}

		teachers = append(teachers, teacher)
	}
	return teachers, nil
}
func generateInsertQuery(model interface{}) string {
	modelType := reflect.TypeOf(model)
	var columns, placeholders string
	for i := 0; i < modelType.NumField(); i++ {
		dbTag := modelType.Field(i).Tag.Get("db")
		fmt.Println("dbTag", dbTag)
		dbTag = strings.TrimSuffix(dbTag, ",omitempty")

		if dbTag != "" && dbTag != "id" {
			if columns != "" {
				columns += ", "
				placeholders += ", "

			}
			columns += dbTag
			placeholders += "?"
		}

	}
	fmt.Printf("INSERT INTO teachers (%s) VALUES (%s)", columns, placeholders)
	return fmt.Sprintf("INSERT INTO teachers (%s) VALUES (%s)", columns, placeholders)
}
func getStructValues(model interface{}) []interface{} {
	modelValue := reflect.ValueOf(model)
	modelType := modelValue.Type()
	values := []interface{}{}
	for i := 0; i < modelType.NumField(); i++ {
		dbTag := modelType.Field(i).Tag.Get("db")
		if dbTag != "" && dbTag != "id,omitempty" {
			values = append(values, modelValue.Field(i).Interface())
		}
	}
	fmt.Printf("%v", values)

	return values

}
func GetOneTeacherDbHandler(w http.ResponseWriter, id int) (models.Teacher, error) {
	db, err := ConnectDb()
	if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Error connecting to database")

	}
	defer db.Close()
	var teacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name,email, class, subject FROM teachers WHERE id=?", id).Scan(&teacher.Id, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
	if err == sql.ErrNoRows {

		http.Error(w, "Teacher not found", http.StatusNotFound)
		return models.Teacher{}, err
	} else if err != nil {
		return models.Teacher{}, utils.ErrorHandler(err, "Database query error")

	}
	return teacher, nil
}

func AddTeacherDbHandler(newTeachers []models.Teacher) ([]models.Teacher, error) {
	db, err := ConnectDb()

	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")

	}
	defer db.Close()
	// stmt, err := db.Prepare("INSERT INTO teachers(first_name, last_name, email, class, subject) VALUES(?, ?, ?, ?,?)")
	stmt, err := db.Prepare(generateInsertQuery(models.Teacher{}))
	if err != nil {

		return nil, utils.ErrorHandler(err, "Error in preparing SQL query")

	}

	defer stmt.Close()

	addedTeachers := make([]models.Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		// res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		values := getStructValues(newTeacher)
		res, err := stmt.Exec(values...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting data into database")

		}
		lastID, err := res.LastInsertId()
		if err != nil {

			return nil, utils.ErrorHandler(err, "Error getting last insert Id")

		}
		newTeacher.Id = int(lastID)
		addedTeachers[i] = newTeacher
	}
	return addedTeachers, nil
}

func UpdateTeacherDbHandler(id int, updatedTeacher models.Teacher) (models.Teacher, error) {
	db, err := ConnectDb()
	if err != nil {

		return models.Teacher{}, utils.ErrorHandler(err, "unable to connect to database id")

	}
	defer db.Close()
	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id=?", id).Scan(&existingTeacher.Id, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Teacher{}, utils.ErrorHandler(err, "Teacher not found")

		}
		return models.Teacher{}, utils.ErrorHandler(err, "Unable to retrieve data")

	}
	updatedTeacher.Id = existingTeacher.Id
	_, err = db.Exec("UPDATE teachers set first_name=?, last_name=?, email=?, class=?, subject=? WHERE id=?", updatedTeacher.FirstName, updatedTeacher.LastName, updatedTeacher.Email, updatedTeacher.Class, updatedTeacher.Subject, updatedTeacher.Id)

	if err != nil {

		return models.Teacher{}, utils.ErrorHandler(err, "Teacher not found")

	}
	return updatedTeacher, err
}

func PatchTeachersDbHandler(updates []map[string]interface{}) error {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		return utils.ErrorHandler(err, "Unable to connect database")

	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return utils.ErrorHandler(err, "Error starting transaction")

	}

	for _, update := range updates {
		idstr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			return utils.ErrorHandler(err, "Invalid Teacher ID in update")

		}

		id, err := strconv.Atoi(idstr)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error while Converting ID to int")

		}

		var teacher models.Teacher
		err = tx.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id=?", id).
			Scan(&teacher.Id, &teacher.FirstName, &teacher.LastName, &teacher.Email, &teacher.Class, &teacher.Subject)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				return utils.ErrorHandler(err, "Teacher not found")

			}
			tx.Rollback()
			return utils.ErrorHandler(err, "Error retrieving teacher")

		}

		teacherVal := reflect.ValueOf(&teacher).Elem()
		teacherType := teacherVal.Type()

		for k, v := range update {
			if k == "id" {
				continue
			}
			for i := 0; i < teacherVal.NumField(); i++ {
				field := teacherType.Field(i)
				tag := strings.Split(field.Tag.Get("json"), ",")[0] // strip omitempty
				if tag == k {
					fieldVal := teacherVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(v)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							return utils.ErrorHandler(err, "Type mismatch in update")

						}
					}
				}
			}
		}

		_, err = tx.Exec("UPDATE teachers SET first_name=?, last_name=?, email=?, class=?, subject=? WHERE id=?",
			teacher.FirstName, teacher.LastName, teacher.Email, teacher.Class, teacher.Subject, teacher.Id)
		if err != nil {
			tx.Rollback()
			return utils.ErrorHandler(err, "Error updating Teacher")

		}
	}

	if err := tx.Commit(); err != nil {
		return utils.ErrorHandler(err, "Error while commiting the transaction")

	}
	return nil
}

func PatchOneTeacherDbHandler(id int, updates map[string]interface{}) (models.Teacher, error) {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		return models.Teacher{}, utils.ErrorHandler(err, "Unable to connect database")

	}
	defer db.Close()
	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id=?", id).Scan(&existingTeacher.Id, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Class, &existingTeacher.Subject)
	if err != nil {
		if err == sql.ErrNoRows {

			return models.Teacher{}, utils.ErrorHandler(err, "Teacher not found")

		}
		return models.Teacher{}, utils.ErrorHandler(err, "Unable to retrieve data")
	}

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

		return models.Teacher{}, utils.ErrorHandler(err, "Teacher not found")

	}
	return existingTeacher, nil
}

func DeleteOneTeacherDbHandler(id int) error {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		return utils.ErrorHandler(err, "Unable to connect to the database")

	}
	defer db.Close()
	result, err := db.Exec("DELETE FROM teachers WHERE id=?", id)
	if err != nil {

		return utils.ErrorHandler(err, "Teacher not found")

	}
	fmt.Println(result.RowsAffected())
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "Error while retrieving delete the data")
	}
	if rowsAffected == 0 {
		// http.Error(w, "Teacher not Found", http.StatusNotFound)
		return err
	}
	return nil
}

func DeleteTeachersDbHandler(ids []int) ([]int, error) {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Unable to connect to the database")

	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Error Starting Transaction")

	}
	stmt, err := tx.Prepare("DELETE FROM teachers WHERE id=?")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return nil, utils.ErrorHandler(err, "Error while preparing Delete SQL Statement ")

	}
	defer stmt.Close()

	deletedIds := []int{}
	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			log.Println(err)
			return nil, utils.ErrorHandler(err, "Error while Deleting Teachers")
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, "Error retrieving deleted result")
		}
		//if teacher was deleted then add the ID to the deletedIds slice
		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)

		}
		if rowsAffected < 1 {
			tx.Rollback()
			return nil, utils.ErrorHandler(err, fmt.Sprintf("Id %d doesn't exixt", id))

		}

	}
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Error while commiting transaction")

	}
	if len(deletedIds) < 1 {
		// http.Error(w, "IDs donot exists", http.StatusBadRequest)
		return nil, utils.ErrorHandler(err, "IDs donot exist")
	}
	return deletedIds, nil
}

func GetStudentByTeacherId(id int) ([]models.Student, error) {
	var students []models.Student

	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "unable to connect db")

	}

	defer db.Close()
	query := `SELECT id, first_name, last_name, email, class FROM students where class=(SELECT class FROM teachers WHERE id=?)`
	rows, err := db.Query(query, id)
	if err != nil {
		log.Println("Error where query")
		return nil, utils.ErrorHandler(err, "unable to fetch student")
	}
	defer rows.Close()

	for rows.Next() {
		var student models.Student
		err = rows.Scan(&student.Id, &student.FirstName, &student.LastName, &student.Email, &student.Class)

		if err != nil {
			log.Println("error while scanning")
			return nil, utils.ErrorHandler(err, "internal error")
		}
		students = append(students, student)
	}
	err = rows.Err()
	if err != nil {
		log.Println("Error while iteration")
		return nil, utils.ErrorHandler(err, "error during fetching ")
	}
	return students, nil
}

func GetStudentCountByTeacherIdDbHandler(id int) (int, error) {

	var studentCount int
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		return 0, utils.ErrorHandler(err, "unable to connect db")

	}

	defer db.Close()
	query := `SELECT COUNT(*) FROM students WHERE class=(SELECT class FROM teachers WHERE id=?)`
	err = db.QueryRow(query, id).Scan(&studentCount)
	if err != nil {
		log.Println("error fetching and scanning count query", err)
		return 0, utils.ErrorHandler(err, "unable to fetch student")
	}

	return studentCount, nil
}
