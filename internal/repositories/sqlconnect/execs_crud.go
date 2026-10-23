package sqlconnect

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"restapi/internal/models"
	"restapi/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-mail/mail/v2"
)

func addExecsFilter(r *http.Request, query string, args []interface{}) (string, []interface{}) {
	params := map[string]string{
		"first_name":      "first_name",
		"last_name":       "last_name",
		"email":           "email",
		"username":        "username",
		"inactive_status": "inactive_status",
		"role":            "role",
	}

	for param, dbField := range params {
		value := r.URL.Query().Get(param)
		if value != "" {
			query += " AND " + dbField + " = ?"
			args = append(args, value)
		}
	}

	return query, args
}

func addExecsSorting(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortby"]
	validSorts := make([]string, 0, len(sortParams))

	for _, param := range sortParams {
		parts := strings.Split(param, ":")
		if len(parts) != 2 {
			continue
		}

		field := parts[0]
		order := strings.ToUpper(parts[1])

		if !isValidExecSortField(field) || !isValidSortOrder(order) {
			continue
		}

		validSorts = append(validSorts, field+" "+order)
	}

	if len(validSorts) > 0 {
		query += " ORDER BY " + strings.Join(validSorts, ", ")
	}

	return query
}

func isValidExecSortField(field string) bool {
	validFields := map[string]bool{
		"first_name":      true,
		"last_name":       true,
		"email":           true,
		"class":           true,
		"username":        true,
		"user_created_at": true,
	}

	return validFields[field]
}

func GetExecsDbHandler(execs []models.Exec, r *http.Request) ([]models.Exec, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE 1=1"
	var args []interface{}

	query, args = addExecsFilter(r, query, args)
	query = addExecsSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Database query error")
	}
	defer rows.Close()

	for rows.Next() {
		var exec models.Exec

		err = rows.Scan(
			&exec.ID,
			&exec.FirstName,
			&exec.LastName,
			&exec.Email,
			&exec.Username,
			&exec.UserCreatedAt,
			&exec.InactiveStatus,
			&exec.Role,
		)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error scanning database results")
		}

		execs = append(execs, exec)
	}

	if err = rows.Err(); err != nil {
		return nil, utils.ErrorHandler(err, "Error reading database results")
	}

	return execs, nil
}

func AddExecsDbHandler(newExecs []models.Exec) ([]models.Exec, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	stmt, err := db.Prepare(
		"INSERT INTO execs (first_name, last_name, email, username, password, inactive_status, role) VALUES (?, ?, ?, ?, ?, ?, ?)",
	)
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error preparing SQL query")
	}
	defer stmt.Close()

	addedExecs := make([]models.Exec, 0, len(newExecs))

	for _, newExec := range newExecs {
		newExec.Password, err = utils.HashPassword(newExec.Password)
		if err != nil {
			return nil, err
		}
		result, err := stmt.Exec(
			newExec.FirstName,
			newExec.LastName,
			newExec.Email,
			newExec.Username,
			newExec.Password,
			newExec.InactiveStatus,
			newExec.Role,
		)
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error inserting exec")
		}

		lastID, err := result.LastInsertId()
		if err != nil {
			return nil, utils.ErrorHandler(err, "Error getting inserted exec ID")
		}

		newExec.ID = int(lastID)
		addedExecs = append(addedExecs, newExec)
	}

	return addedExecs, nil
}

func GetOneExecDbHandler(id int) (models.Exec, error) {
	db, err := ConnectDb()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	var exec models.Exec

	query := "SELECT id, first_name, last_name, email, username, role, inactive_status, user_created_at FROM execs WHERE id = ?"

	err = db.QueryRow(query, id).Scan(
		&exec.ID,
		&exec.FirstName,
		&exec.LastName,
		&exec.Email,
		&exec.Username,
		&exec.Role,
		&exec.InactiveStatus,
		&exec.UserCreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Exec{}, fmt.Errorf("exec with id %d not found", id)
		}

		return models.Exec{}, utils.ErrorHandler(err, "Error retrieving exec")
	}

	return exec, nil
}

func DeleteOneExecDbHandler(id int) error {
	db, err := ConnectDb()
	if err != nil {
		return utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM execs WHERE id = ?", id)
	if err != nil {
		return utils.ErrorHandler(err, "Error deleting exec")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return utils.ErrorHandler(err, "Error checking deleted exec")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("exec with id %d not found", id)
	}

	return nil
}

func PatchOneExecDbHandler(id int, payload map[string]interface{}) (models.Exec, error) {
	if len(payload) == 0 {
		return models.Exec{}, fmt.Errorf("payload empty")
	}

	columns := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"username":   "username",
	}

	setParts := make([]string, 0, len(payload))
	args := make([]interface{}, 0, len(payload))

	for key, value := range payload {
		column, ok := columns[key]
		if !ok {
			return models.Exec{}, fmt.Errorf("field %q cannot be updated", key)
		}

		setParts = append(setParts, column+" = ?")
		args = append(args, value)
	}

	query := "UPDATE execs SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	args = append(args, id)

	db, err := ConnectDb()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	result, err := db.Exec(query, args...)
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error updating exec")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Exec{}, utils.ErrorHandler(err, "Error checking updated exec")
	}

	var exec models.Exec

	selectQuery := "SELECT id, first_name, last_name, email, username, inactive_status, user_created_at, role FROM execs WHERE id = ?"

	err = db.QueryRow(selectQuery, id).Scan(
		&exec.ID,
		&exec.FirstName,
		&exec.LastName,
		&exec.Email,
		&exec.Username,
		&exec.InactiveStatus,
		&exec.UserCreatedAt,
		&exec.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Exec{}, fmt.Errorf("exec with id %d not found", id)
		}

		return models.Exec{}, utils.ErrorHandler(err, "Error retrieving updated exec")
	}

	if rowsAffected == 0 {
		return exec, nil
	}

	return exec, nil
}

func PatchExecDbHandler(payloads []map[string]interface{}) ([]models.Exec, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, utils.ErrorHandler(err, "Error starting transaction")
	}
	defer tx.Rollback()

	columns := map[string]string{
		"first_name": "first_name",
		"last_name":  "last_name",
		"email":      "email",
		"username":   "username",
	}

	updatedExecs := make([]models.Exec, 0, len(payloads))

	for _, payload := range payloads {
		if len(payload) == 0 {
			return nil, fmt.Errorf("payload empty")
		}

		idValue, ok := payload["id"]
		if !ok {
			return nil, fmt.Errorf("missing required id field")
		}

		id, ok := idValue.(int)
		if !ok {
			return nil, fmt.Errorf("id must be an integer")
		}

		setParts := make([]string, 0, len(payload))
		args := make([]interface{}, 0, len(payload))

		for key, value := range payload {
			if key == "id" {
				continue
			}

			column, ok := columns[key]
			if !ok {
				return nil, fmt.Errorf("field %q cannot be updated", key)
			}

			setParts = append(setParts, column+" = ?")
			args = append(args, value)
		}

		if len(setParts) > 0 {
			query := "UPDATE execs SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
			args = append(args, id)

			_, err = tx.Exec(query, args...)
			if err != nil {
				return nil, utils.ErrorHandler(err, "Error updating exec")
			}
		}

		var exec models.Exec

		query := "SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE id = ?"

		err = tx.QueryRow(query, id).Scan(
			&exec.ID,
			&exec.FirstName,
			&exec.LastName,
			&exec.Email,
			&exec.Username,
			&exec.UserCreatedAt,
			&exec.InactiveStatus,
			&exec.Role,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("exec with id %d not found", id)
			}

			return nil, utils.ErrorHandler(err, "Error retrieving updated exec")
		}

		updatedExecs = append(updatedExecs, exec)
	}

	if err = tx.Commit(); err != nil {
		return nil, utils.ErrorHandler(err, "Error committing transaction")
	}

	return updatedExecs, nil
}

func GetUserByUsername(username string) (*models.Exec, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "database error")
	}
	defer db.Close()
	user := &models.Exec{}
	err = db.QueryRow(`SELECT id, first_name, last_name, email, username, password, inactive_status, role FROM execs WHERE username=?`, username).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.Password, &user.InactiveStatus, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrorHandler(err, "user not found")

		}
		return nil, utils.ErrorHandler(err, "database error")
	}
	return user, nil
}
func UpdatePassworDdbHandler(userId int, req models.UpdatePasswordRequest) (bool,error) {
	db, err := ConnectDb()
	if err != nil {
		
		return false, utils.ErrorHandler(err, "database connection error")
	}
	defer db.Close()
	var username string
	var userPassword string
	var userRole string
	err = db.QueryRow("SELECT username, password, role FROM execs WHERE id=?", userId).Scan(&username, &userPassword, &userRole)

	if err != nil {
		
		return false, utils.ErrorHandler(err, "user not found")
	}
	err = utils.VerfiyPassword(req.CurrentPassword, userPassword)

	if err != nil {
		
		return false, utils.ErrorHandler(err, "The password you entered does not match the current password on file.")
	}
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		
		return false, utils.ErrorHandler(err, "internal error")
	}

	currentTime := time.Now().Format(time.RFC3339)
	_, err = db.Exec("UPDATE execs SET password=?, password_changed_at=? WHERE id=?", hashedPassword, currentTime, userId)
	if err != nil {

		return false, utils.ErrorHandler(err, "failed to update the password")
	}
	return true, nil
}





func ForgetPasswordDbHandler(emailId string) error {
	db, err := ConnectDb()
	if err != nil {
		return utils.ErrorHandler(err, "Internal error")
		
	}
	defer db.Close()
	var exec models.Exec
	err = db.QueryRow("SELECT id FROM execs WHERE email=?", emailId).Scan(&exec.ID)
	if err != nil {
		return utils.ErrorHandler(err, "User not found")
	
	}
	duration, err := strconv.Atoi(os.Getenv("RESET_TOKEN_EXP_DURATION"))
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")
	
	}
	mins := time.Duration(duration)
	// expiry:=time.Now().Add(mins*time.Minute).Format(time.RFC3339)

	expiry := time.Now().Add(mins * time.Minute).Format("2006-01-02 15:04:05")

	tokenBytes := make([]byte, 32)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")
	
	}
	log.Println("tokenBytes:", tokenBytes)

	token := hex.EncodeToString(tokenBytes)
	hashedToken := sha256.Sum256(tokenBytes)
	log.Println("hashedToken", hashedToken)
	hashedTokenString := hex.EncodeToString(hashedToken[:])
	_, err = db.Exec("UPDATE execs SET password_reset_token=?, password_token_expires=? WHERE id=?", hashedTokenString, expiry, exec.ID)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")

		
	}

	//send the reset email
	resetURL := fmt.Sprintf("https://localhost:3000/execs/resetPassword/reset/%s", token)

	message := fmt.Sprintf("Forgot your password? Reset your password using the following link: \n%s\n if you didn't request a password reset, please ignore this email. This link is only valid for %d minute", resetURL, int(mins))

	m := mail.NewMessage()
	m.SetHeader("From", "paudelchetan1112@gmail.com")
	m.SetHeader("To", emailId)
	m.SetHeader("subject", "Your Password Reset Link")
	m.SetBody("text/plain", message)
	d := mail.NewDialer("localhost", 1025, "", "")
	err = d.DialAndSend(m)
	if err != nil {
		return utils.ErrorHandler(err, "Failed to send password reset email")
		
	}
	fmt.Println(d)
	return nil
}



