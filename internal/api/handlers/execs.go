package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"restapi/internal/models"
	"restapi/internal/repositories/sqlconnect"
	"restapi/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

// ValidateRequired mimics your structural validation helper for an executive.
func Validate_Required(exec models.Exec) error {
	if strings.TrimSpace(exec.FirstName) == "" {
		return errors.New("first_name is required")
	}
	if strings.TrimSpace(exec.LastName) == "" {
		return errors.New("last_name is required")
	}
	if strings.TrimSpace(exec.Email) == "" {
		return errors.New("email is required")
	}
	if strings.TrimSpace(exec.Username) == "" {
		return errors.New("username is required")
	}
	if strings.TrimSpace(exec.Password) == "" {
		return errors.New("password is required")
	}
	if strings.TrimSpace(exec.Role) == "" {
		return errors.New("role is required")
	}
	return nil
}

// GetExecsHandler fetches all executive records.
func GetExecsHandler(w http.ResponseWriter, r *http.Request) {
	var execs []models.Exec
	// Fixed the syntax assignment typo from your snippet
	var err error
	execs, err = sqlconnect.GetExecsDbHandler(execs, r)
	if err != nil {
		log.Printf("Error fetching executives: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "Success",
		Count:  len(execs),
		Data:   execs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddExecsHandler securely intercepts payloads and creates executive records.
func AddExecsHandler(w http.ResponseWriter, r *http.Request) {
	var newExecs []models.Exec
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&newExecs)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	// Validate fields and securely hash passwords before passing down to db handler
	for i := range newExecs {
		if err := Validate_Required(newExecs[i]); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	addedExecs, err := sqlconnect.AddExecsDbHandler(newExecs)
	if err != nil {
		log.Printf("Error during batch record creation: %v", err)

		// Catch common duplicate key syntax issues cleanly
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint") {
			http.Error(w, "A user with this email or username already exists", http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(addedExecs),
		Data:   addedExecs,
	}
	json.NewEncoder(w).Encode(response)
}

// GetOneExecsHandler loads a single targeted executive record.
func GetOneExecsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	exec, err := sqlconnect.GetOneExecDbHandler(id)
	if err != nil {
		log.Printf("Error pulling profile item %d: %v", id, err)
		if strings.Contains(err.Error(), "no rows in result set") {
			http.Error(w, "Executive record not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string      `json:"status"`
		Data   models.Exec `json:"data"`
	}{
		Status: "success",
		Data:   exec,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteOneExecsHandler cleanly drops an entry from database state.
func DeleteOneExecsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DeleteOneExecDbHandler(id)
	if err != nil {
		log.Printf("Error deleting user %d: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status        string `json:"status"`
		DeletedExecId int    `json:"deleted_exec_id"` // Adjusted to reflect entity structural naming
	}{
		Status:        "Deleted successfully",
		DeletedExecId: id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PatchOneExecsHandler filters inputs and safely applies a delta profile update.
func PatchOneExecsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("Invalid input format for path variable ID")
		http.Error(w, "Invalid path parameters", http.StatusBadRequest)
		return
	}

	var payload map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Println("Error while parsing incoming JSON input stream")
		http.Error(w, "Invalid body JSON format", http.StatusBadRequest)
		return
	}

	// Safety: Remove properties that are auto-managed or high-privilege
	delete(payload, "id")
	delete(payload, "role")
	delete(payload, "inactive_status")

	// If security actions or password patches occur, they must be processed via their own strict methods
	if rawPassword, ok := payload["password"].(string); ok {
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error parsing secure update data", http.StatusInternalServerError)
			return
		}
		payload["password"] = string(hashedBytes)
	}

	if len(payload) == 0 {
		http.Error(w, "No valid or permissible fields provided for patch update", http.StatusBadRequest)
		return
	}

	updatedExec, err := sqlconnect.PatchOneExecDbHandler(id, payload)
	if err != nil {
		log.Printf("Error updating target row %d: %v", id, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status string      `json:"status"`
		Data   models.Exec `json:"data"`
	}{
		Status: "Update Successful",
		Data:   updatedExec,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func PatchExecsHandler(w http.ResponseWriter, r *http.Request) {
	var payload []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Println("Error parsing stream content payload array block")
		http.Error(w, "Malformed array configuration content", http.StatusBadRequest)
		return
	}

	for _, item := range payload {
		delete(item, "role")
		delete(item, "inactive_status")
		if rawPassword, ok := item["password"].(string); ok {
			hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
			item["password"] = string(hashedBytes)
		}
	}

	updatedExecs, err := sqlconnect.PatchExecDbHandler(payload)
	if err != nil {
		log.Printf("Error executing collection patching: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Status       string        `json:"status"`
		Count        int           `json:"count"`
		UpdatedExecs []models.Exec `json:"updated_execs"`
	}{
		Status:       "Successfully updated",
		Count:        len(updatedExecs),
		UpdatedExecs: updatedExecs,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Exec

	//Data validation
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid Request Body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	//Search for user if user actually exists

	user, err := sqlconnect.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "invalid username or password", http.StatusBadRequest)
		return
	}

	//is user active
	if user.InactiveStatus {
		http.Error(w, "Account is inactive", http.StatusForbidden)
		return
	}

	//verify password
	err = utils.VerfiyPassword(req.Password, user.Password)

	if err != nil {
		http.Error(w, "could not create login token", http.StatusInternalServerError)
	}

	//generate Token
	tokenString, err := utils.SignToken(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	})

	//send token as a response or as a cookie

	response := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteStrictMode,
	})
	w.Header().Set("content-type", "application/json")
	w.Write([]byte(`{"message":"Logout Successfully"}`))
}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid exec ID", http.StatusBadRequest)
		return
	}
	var req models.UpdatePasswordRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	r.Body.Close()
	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "Please enter password", http.StatusBadRequest)
		return
	}
	_, err = sqlconnect.UpdatePassworDdbHandler(userId, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// token, err:=utils.SignToken(userId, username, userRole)
	// if err!=nil{
	// 	utils.ErrorHandler(err, "Password updated. could not create token")
	// 	return
	// }

	// 	http.SetCookie(w, &http.Cookie{
	// 		Name:     "Bearer",
	// 		Value:    token,
	// 		Path:     "/",
	// 		HttpOnly: true,
	// 		Secure:   true,
	// 		Expires:  time.Now().Add(24 * time.Hour),
	// 		SameSite: http.SameSiteStrictMode,
	// 	})

	//send token as a response or as a cookie

	response := struct {
		Message string `json:"message"`
	}{
		Message: "Password updated successful",
	}
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	error := sqlconnect.ForgetPasswordDbHandler(req.Email)
	if error != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
 
}


	
	


// Request defines the expected JSON payload
type ResetRequest struct {
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	// 1. Extract the token from the path (Go 1.22+ routing)
	token := r.PathValue("resetcode")
	if strings.TrimSpace(token) == "" {
		http.Error(w, "Reset token is required", http.StatusBadRequest)
		return
	}

	// 2. Decode the incoming JSON payload (Fix: Must pass pointer &req)
	var req ResetRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// 3. Data validation for blank values
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	req.ConfirmPassword = strings.TrimSpace(req.ConfirmPassword)

	if req.NewPassword == "" || req.ConfirmPassword == "" {
		http.Error(w, "Passwords cannot be blank", http.StatusBadRequest)
		return
	}

	// Ensure passwords match
	if req.NewPassword != req.ConfirmPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	// Optional: Add password strength checking here (e.g., len(req.NewPassword) < 8)
	if len(req.NewPassword) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// 4. Connect to the database
	db, err := sqlconnect.ConnectDb()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// Note: If ConnectDb returns a pooled connection, do not blindly close it. 
	// If it opens a raw fresh connection each time, defer db.Close() here.

	// 5. Query user details by token and check expiration
	var userID int
	var storedTokenHash string
	var expiresAt time.Time

	// Adjust this SQL query to match your schema (e.g., users table or password_resets table)
	query := "SELECT id, password_reset_token, password_token_expires FROM execs WHERE reset_token_hash = ?"
	
	// Assuming token is securely hashed in the DB. If storing raw tokens, query directly.
	// If you use a hashing mechanism for tokens, hash `token` before querying.
	err = db.QueryRow(query, token).Scan(&userID, &storedTokenHash, &expiresAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if token has expired
	if time.Now().After(expiresAt) {
		http.Error(w, "Reset token has expired", http.StatusUnauthorized)
		return
	}

	// 6. Securely hash the new password using bcrypt
	
	hashedPassword, err := utils.HashPassword(req.ConfirmPassword)
	if err != nil {
		http.Error(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// 7. Update password and invalidate the token so it can't be reused
	updateQuery := `
		UPDATE execs 
		SET password = $1, 
		    password_reset_token= NULL, 
		     password_token_expires= NULL 
		WHERE id = $2`

	_, err = db.Exec(updateQuery, string(hashedPassword), userID)
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	// 8. Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successfully"})
}
