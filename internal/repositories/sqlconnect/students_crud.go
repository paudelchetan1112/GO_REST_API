package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"restapi/internal/models"
	"restapi/internal/repositories/sqlconnect/dbutil"
	"restapi/pkg/utils"

	"strings"

	
)

func addStudentFilter(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"class":      "class",
	}
	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + dbField + "=?"
			args = append(args, value)
		}
	}
	return query, args
}
func addStudentSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		query += " ORDER BY"
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidStudentSortField(field) || !isValidStudentSortOrder(order) {
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
func isValidStudentSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}
func isValidStudentSortField(field string) bool {
	validField := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
	}
	return validField[field]

}
func GetStudentsDbHandler(students []models.Student, r *http.Request) ([]models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class FROM students WHERE 1=1"
	var args []interface{}
	query, args = addStudentFilter(r, query, args)
	query = addStudentSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println("Error", err)
		return nil, utils.ErrorHandler(err, "Database Query error")

	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, utils.ErrorHandler(err, "Something went wrong")
	}
	for rows.Next() {
		var student models.Student
		err = rows.Scan(&student.Id, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error scanning database results")

		}
		students = append(students, student)

	}
	return students, nil

}

func AddStudentDbHandler(newStudents []models.Student) ([]models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to the database")

	}
	defer db.Close()
	stmt, err := db.Prepare("INSERT INTO students(first_name, last_name, email, class) VALUES(?, ?, ?, ?)")
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error in preparing SQL query")

	}
	defer stmt.Close()
	addedStudents := make([]models.Student, len(newStudents))
	for i, newStudent := range newStudents {
		value := dbutil.GetStructValues(newStudent)
		fmt.Println(value...)
		result, err := stmt.Exec(value...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting data into databases")
		}
		lastID, err := result.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error getting last insert Id")

		}
		newStudent.Id = int(lastID)
		addedStudents[i] = newStudent
	}
	return addedStudents, nil
}
func GetOneStudentDbHandler(id int) (models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		return models.Student{}, err
	}
	defer db.Close()
	var student models.Student
	query := "SELECT id, first_name, last_name, email, class FROM students WHERE id=?"
	db.QueryRow(query, id).Scan(&student.Id, &student.FirstName, &student.LastName, &student.Email, &student.Class)

	return student, nil

}
func DeleteOneStudentDbHandler(id int) error {

	db, err := ConnectDb()
	if err != nil {
		return err
	}
	defer db.Close()
	stmt, err := db.Prepare("DELETE FROM students WHERE id=?")
	if err != nil {
		return err
	}
	result, err := stmt.Exec(id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%v", "Student not found")
	}

	return nil
}

func DeleteStudentDbHandler(ids []int) ([]int, error) {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Unable to connect to the database")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Error starting transaction")
	}
	// Defensively roll back if we return early due to errors before committing
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM students WHERE id=?")
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Error while preparing Delete SQL statement")
	}
	defer stmt.Close()

	deletedIds := []int{}
	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			log.Println(err)
			return nil, utils.ErrorHandler(err, "Error while deleting Student")
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error retrieving deleted result")
		}

		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)
		} else {
			// If you want the entire batch to fail when ONE ID is missing, keep this.
			// If you want to skip missing IDs and delete the rest, delete these 2 lines.
			err = fmt.Errorf("id %d doesn't exist", id)
			return nil, utils.ErrorHandler(err, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err)
		return nil, utils.ErrorHandler(err, "Error while committing transaction")
	}

	if len(deletedIds) < 1 {
		err = fmt.Errorf("no IDs were found to delete")
		return nil, utils.ErrorHandler(err, "IDs do not exist")
	}

	return deletedIds, nil
}

func UpdateStudentDbHandler(id int, student models.Student) (models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error while connecting db")

	}
	defer db.Close()
	stmt, err := db.Prepare("UPDATE students SET first_name=?, last_name=?, email=?, class=? WHERE id=?")
	if err != nil {
		log.Println(err)
		return models.Student{}, utils.ErrorHandler(err, "Error while preparing db")

	}
	defer stmt.Close()
	result, err := stmt.Exec(student.FirstName, student.LastName, student.Email, student.Class, id)
	if err != nil {
		log.Println(err)
		return models.Student{}, utils.ErrorHandler(err, "Internal Error")

	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		return models.Student{}, utils.ErrorHandler(err, "Internal Error")
	}
	if rowsAffected < 1 {
		log.Println("student of that Id is not found")

		return models.Student{}, fmt.Errorf("Student not found")

	}

	return student, nil
}

func PatchOneStudentDbHandler(id int, payload map[string]interface{}) (models.Student, error) {
	query := "UPDATE students SET "
	var args []interface{}
	if len(payload) == 0 {
		return models.Student{}, fmt.Errorf("payload empty")
	}

	for key, value := range payload {
		var column string
		switch key {
		case "first_name":
			column = "first_name"
		case "last_name":
			column = "last_name"
		case "email":
			column = "email"
		case "class":
			column = "class"
		default:
			continue
		}
		query += column + "=?, "
		args = append(args, value)

	}
	if len(args) == 0 {
		return models.Student{}, fmt.Errorf("no valid fields provided for update")
	}
	query = query[:len(query)-2]

	query += " WHERE id=?"
	args = append(args, id)
	db, err := ConnectDb()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Error While connecting db")

	}
	defer db.Close()
	stmt, err := db.Prepare(query)
	if err != nil {

		return models.Student{}, utils.ErrorHandler(err, "Error while preparing query")

	}
	defer stmt.Close()
	result, err := stmt.Exec(args...)
 if err!=nil {
		return models.Student{}, utils.ErrorHandler(err, "Error while Executing command")

	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "Internal Server Error ")
	}
	if rowsAffected == 0 {
		var exists int
		// Double-check if the unique record actually lives inside the engine row indexes
		checkErr := db.QueryRow("SELECT 1 FROM students WHERE id = ?", id).Scan(&exists)
		if checkErr == sql.ErrNoRows {
			return models.Student{}, fmt.Errorf("Student not found") // Genuine 404 condition
		} else if checkErr != nil {
			return models.Student{}, utils.ErrorHandler(checkErr, "Database integrity lookup error")
		}
		// If checkErr == nil, it means the record is real, but values were identical!
		// We do not throw an error here. We fall through to return the clean structural response.
		log.Println("Patch payload matched active database states identically.")
	}
	var Student models.Student

	err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(&Student.Id, &Student.FirstName, &Student.LastName, &Student.Email, &Student.Class)
	if err != nil {
		return models.Student{}, utils.ErrorHandler(err, "database error")

	}
	return Student, nil

}


// PatchStudentDbHandler updates multiple students in one transaction and returns their updated states.
func PatchStudentDbHandler(payloads []map[string]interface{}) ([]models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error While connecting db")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "unable to initialize transaction")
	}
	defer tx.Rollback()

	var updatedStudents []models.Student

	for _, payload := range payloads {
		if len(payload) == 0 {
			return nil, fmt.Errorf("payload empty")
		}

		// Extract target ID from map
		id, ok := payload["id"]
		if !ok {
			return nil, fmt.Errorf("missing required 'id' field for student update")
		}

		query := "UPDATE students SET "
		var args []interface{}

		for key, value := range payload {
			var column string
			switch key {
			case "first_name":
				column = "first_name"
			case "last_name":
				column = "last_name"
			case "email":
				column = "email"
			case "class":
				column = "class"
			default:
				continue
			}
			query += column + "=?, "
			args = append(args, value)
		}

		// If no fields besides ID were provided, skip execution but still track the student
		if len(args) == 0 {
			var student models.Student
			err = tx.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).
				Scan(&student.Id, &student.FirstName, &student.LastName, &student.Email, &student.Class)
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("student with id %v not found", id)
			} else if err != nil {
				return nil, utils.ErrorHandler(err, "Failed to retrieve student state")
			}
			updatedStudents = append(updatedStudents, student)
			continue
		}

		query = query[:len(query)-2]
		query += " WHERE id=?"
		args = append(args, id)

		result, err := tx.Exec(query, args...)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error while Executing command")
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Internal Server Error")
		}

		// Handle when no rows are modified (either bad ID or identical data sent)
		if rowsAffected == 0 {
			var exists int
			checkErr := tx.QueryRow("SELECT 1 FROM students WHERE id = ?", id).Scan(&exists)
			if checkErr == sql.ErrNoRows {
				return nil, fmt.Errorf("student with id %v not found", id)
			} else if checkErr != nil {
				return nil, utils.ErrorHandler(checkErr, "Database error during existence verification")
			}
		}

		// Fetch the freshly updated state of this student inside the loop
		var student models.Student
		err = tx.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).
			Scan(&student.Id, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Failed to retrieve database state")
		}

		updatedStudents = append(updatedStudents, student)
	}

	if err := tx.Commit(); err != nil {
		return nil, utils.ErrorHandler(err, "Failed to commit batch transaction updates")
	}

	return updatedStudents, nil
}
