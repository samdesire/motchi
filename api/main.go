// Package main implements a backend server for a game application.
// It provides OAuth2-based authentication, WebSocket support for real-time communication,
// and REST APIs for user and pet management.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type PetMoneyUpdate struct {
	PetID  int `json:"pet_id"`
	Amount int `json:"amount"`
}

var (
	db                *sql.DB
	oauth2Server      *server.Server
	oauthClientID     string
	oauthClientSecret string
	upgrader          = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for simplicity
		},
	}
	connections   = make(map[int]*websocket.Conn) // Map of user ID to WebSocket connection
	connectionsMu sync.Mutex                      // Mutex to protect the connections map
	logLevel      string
)

// hashPassword hashes a plaintext password using bcrypt.
// Parameters:
// - password: The plaintext password to hash.
// Returns:
// - The hashed password as a string.
// - An error if hashing fails.
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// validateUserForeignKeys checks if a user has valid foreign key references for their pet and significant other.
// Parameters:
// - userID: The ID of the user to validate.
// Returns:
// - An error if the foreign keys are invalid or if the query fails.
func validateUserForeignKeys(userID int) error {
	var petID, soID sql.NullInt64
	err := db.QueryRow("SELECT pet_id, SO FROM users WHERE id = ?", userID).Scan(&petID, &soID)
	if err != nil {
		return err
	}
	if !petID.Valid || !soID.Valid {
		return fmt.Errorf("invalid access: user %d has invalid foreign keys", userID)
	}
	return nil
}

// getOtherOwner retrieves the ID of the other owner of a pet.
// Parameters:
// - petID: The ID of the pet.
// - currentOwnerID: The ID of the current owner making the request.
// Returns:
// - The ID of the other owner as sql.NullInt64.
// - An error if the query fails.
func getOtherOwner(petID int, currentOwnerID int) (sql.NullInt64, error) {
	var otherOwnerID sql.NullInt64
	err := db.QueryRow("SELECT CASE WHEN main_owner = ? THEN owner2 ELSE main_owner END FROM pets WHERE id = ?", currentOwnerID, petID).Scan(&otherOwnerID)
	if err != nil {
		return sql.NullInt64{}, err
	}
	return otherOwnerID, nil
}

// validateAndUpdatePetMoney validates and updates the money attribute of a pet.
// Parameters:
// - petID: The ID of the pet.
// - amount: The amount to update the pet's money by.
// Returns:
// - A boolean indicating if the update was valid.
// - The new money value after the update.
// - An error if the query fails.
func validateAndUpdatePetMoney(petID int, amount int) (bool, int, error) {
	var currentMoney int
	err := db.QueryRow("SELECT money FROM pets WHERE id = ?", petID).Scan(&currentMoney)
	if err != nil {
		return false, 0, err
	}

	if amount > currentMoney {
		return false, currentMoney, nil
	}

	newMoney := currentMoney - amount
	_, err = db.Exec("UPDATE pets SET money = ? WHERE id = ?", newMoney, petID)
	if err != nil {
		return false, 0, err
	}

	return true, newMoney, nil
}

func logMessage(event string, details map[string]interface{}) {
	if logLevel == "development" {
		log.Printf("Event: %s, Details: %v", event, details)
	} else if logLevel == "production" {
		switch event {
		case "user_login", "create_pet", "add_co_owner":
			log.Printf("Event: %s, Details: %v", event, details)
		}
	}
}

// sendPingMessages periodically sends ping messages to keep the WebSocket connection alive.
// Parameters:
// - userID: The ID of the user associated with the WebSocket connection.
func sendPingMessages(userID int) {
	pingTicker := time.NewTicker(60 * time.Second)
	defer pingTicker.Stop()

	for range pingTicker.C {
		connectionsMu.Lock()
		if conn, ok := connections[userID]; ok {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logMessage("ping_failed", map[string]interface{}{"user_id": userID, "error": err.Error()})
				conn.Close()
				delete(connections, userID)
			}
		}
		connectionsMu.Unlock()
	}
}

// websocketHandler handles WebSocket connections for real-time communication.
// Endpoint: GET /ws
// Behavior:
// - Authenticates the user using an OAuth2 token.
// - Establishes a WebSocket connection.
// - Handles incoming messages and sends responses.
// - Sends periodic ping messages to keep the connection alive.
func websocketHandler(w http.ResponseWriter, r *http.Request) {
	token, err := oauth2Server.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userIDStr := token.GetUserID()
	if userIDStr == "" {
		// Token does not carry a user id (likely client_credentials). Require user-scoped token.
		http.Error(w, "Token must be issued with a user id (use password grant)", http.StatusUnauthorized)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := validateUserForeignKeys(userID); err != nil {
		http.Error(w, "User foreign keys are not valid", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logMessage("ws_upgrade_failed", map[string]interface{}{"error": err.Error()})
		return
	}
	defer conn.Close()

	connectionsMu.Lock()
	connections[userID] = conn
	connectionsMu.Unlock()

	defer func() {
		connectionsMu.Lock()
		if conn, ok := connections[userID]; ok {
			conn.Close()
			delete(connections, userID)
		}
		connectionsMu.Unlock()
	}()

	conn.SetPongHandler(func(appData string) error {
		logMessage("ws_pong", map[string]interface{}{"user_id": userID})
		return nil
	})

	go sendPingMessages(userID)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			logMessage("ws_read_error", map[string]interface{}{"error": err.Error()})
			break
		}

		// Process the message
		logMessage("ws_message_received", map[string]interface{}{"message": string(message)})

		// Determine message type quickly by peeking at the 'type' field
		var msgType struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal(message, &msgType)

		// Handle GetData request: return the caller's associated pet data
		if strings.EqualFold(msgType.Type, "get_data") || strings.EqualFold(msgType.Type, "GetData") {
			// Find the user's pet id (server-sourced)
			var userPetID sql.NullInt64
			if err := db.QueryRow("SELECT pet_id FROM users WHERE id = ?", userID).Scan(&userPetID); err != nil && err != sql.ErrNoRows {
				logMessage("pet_data_error", map[string]interface{}{"error": err.Error(), "user_id": userID})
				conn.WriteJSON(map[string]interface{}{
					"type":    "PetDataResponse",
					"status":  "fail",
					"message": "Server error retrieving pet data",
				})
				continue
			}

			petIDToUse := 0
			if userPetID.Valid {
				petIDToUse = int(userPetID.Int64)
			} else {
				// Check if user is owner2
				var petIDFromPets int
				err := db.QueryRow("SELECT id FROM pets WHERE owner2 = ?", userID).Scan(&petIDFromPets)
				if err != nil {
					if err == sql.ErrNoRows {
						conn.WriteJSON(map[string]interface{}{
							"type":    "PetDataResponse",
							"status":  "fail",
							"message": "Caller has no pet",
						})
						continue
					}
					logMessage("pet_data_error", map[string]interface{}{"error": err.Error(), "user_id": userID})
					conn.WriteJSON(map[string]interface{}{
						"type":    "PetDataResponse",
						"status":  "fail",
						"message": "Server error retrieving pet data",
					})
					continue
				}
				petIDToUse = petIDFromPets
			}

			// Query pet data
			var pet struct {
				ID        int
				Money     int
				Health    int
				Hunger    int
				Happiness int
				MainOwner int `db:"main_owner"`
				Owner2    sql.NullInt64
			}
			row := db.QueryRow("SELECT id, money, health, hunger, happiness, main_owner, owner2 FROM pets WHERE id = ?", petIDToUse)
			if err := row.Scan(&pet.ID, &pet.Money, &pet.Health, &pet.Hunger, &pet.Happiness, &pet.MainOwner, &pet.Owner2); err != nil {
				if err == sql.ErrNoRows {
					conn.WriteJSON(map[string]interface{}{
						"type":    "PetDataResponse",
						"status":  "fail",
						"message": "Pet not found",
					})
					continue
				}
				logMessage("pet_data_error", map[string]interface{}{"error": err.Error(), "pet_id": petIDToUse})
				conn.WriteJSON(map[string]interface{}{
					"type":    "PetDataResponse",
					"status":  "fail",
					"message": "Server error retrieving pet data",
				})
				continue
			}

			// Build response object
			petResp := map[string]interface{}{
				"type":   "PetDataResponse",
				"status": "success",
				"pet": map[string]interface{}{
					"id":         pet.ID,
					"money":      pet.Money,
					"health":     pet.Health,
					"hunger":     pet.Hunger,
					"happiness":  pet.Happiness,
					"main_owner": pet.MainOwner,
					"owner2":     nil,
				},
			}
			if pet.Owner2.Valid {
				petResp["pet"].(map[string]interface{})["owner2"] = int(pet.Owner2.Int64)
			}

			conn.WriteJSON(petResp)
			continue
		}
		// Notify other owner if applicable
		var updateData PetMoneyUpdate
		if err := json.Unmarshal(message, &updateData); err == nil {
			// Derive the pet ID from server-side state (ignore client-supplied pet_id).
			// First, try the user's pet_id column. If not present, check if the user
			// is a co-owner (owner2) in the pets table.
			var userPetID sql.NullInt64
			err := db.QueryRow("SELECT pet_id FROM users WHERE id = ?", userID).Scan(&userPetID)
			if err != nil && err != sql.ErrNoRows {
				logMessage("pet_money_error", map[string]interface{}{"error": err.Error(), "user_id": userID})
				conn.WriteJSON(map[string]interface{}{
					"type":    "ResultResponse",
					"status":  "fail",
					"message": "Server error occurred",
				})
				continue
			}

			petIDToUse := 0
			if userPetID.Valid {
				petIDToUse = int(userPetID.Int64)
			} else {
				// Try to find a pet where this user is owner2
				var petIDFromPets int
				err = db.QueryRow("SELECT id FROM pets WHERE owner2 = ?", userID).Scan(&petIDFromPets)
				if err != nil {
					if err == sql.ErrNoRows {
						conn.WriteJSON(map[string]interface{}{
							"type":    "ResultResponse",
							"status":  "fail",
							"message": "Caller has no pet to operate on",
						})
						continue
					}
					logMessage("pet_money_error", map[string]interface{}{"error": err.Error(), "user_id": userID})
					conn.WriteJSON(map[string]interface{}{
						"type":    "ResultResponse",
						"status":  "fail",
						"message": "Server error occurred",
					})
					continue
				}
				petIDToUse = petIDFromPets
			}

			// Override whatever the client sent and use the server-derived pet id.
			updateData.PetID = petIDToUse

			valid, newMoney, err := validateAndUpdatePetMoney(updateData.PetID, updateData.Amount)
			if err != nil {
				logMessage("pet_money_error", map[string]interface{}{"error": err.Error(), "pet_id": updateData.PetID})
				conn.WriteJSON(map[string]interface{}{
					"type":    "ResultResponse",
					"status":  "fail",
					"message": "Server error occurred",
				})
				continue
			}

			if !valid {
				conn.WriteJSON(map[string]interface{}{
					"type":    "ResultResponse",
					"status":  "fail",
					"message": "Insufficient funds. Pet money cannot go below 0.",
				})
				continue
			}

			conn.WriteJSON(map[string]interface{}{
				"type":     "ResultResponse",
				"status":   "success",
				"newMoney": newMoney,
			})

			otherOwnerID, err := getOtherOwner(updateData.PetID, userID)
			if err == nil && otherOwnerID.Valid {
				connectionsMu.Lock()
				if otherConn, ok := connections[int(otherOwnerID.Int64)]; ok {
					// Broadcast the original message but annotate pet_id with the server-derived value
					// so the recipient sees the authoritative pet id.
					annotated := map[string]interface{}{}
					_ = json.Unmarshal(message, &annotated)
					annotated["pet_id"] = updateData.PetID
					if err = otherConn.WriteJSON(annotated); err != nil {
						logMessage("ws_send_error", map[string]interface{}{"error": err.Error()})
					}
				}
				connectionsMu.Unlock()
			}
		}
	}
}

// createUserHandler handles the creation of a new user account.
// Endpoint: POST /create_user
// Request Body:
// - username: The username of the new user.
// - password: The plaintext password of the new user.
// Response:
// - 201 Created on success.
// - 400 Bad Request if the request body is invalid.
// - 500 Internal Server Error if user creation fails.
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	type CreateUserRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Updated INSERT statement to include name
	_, err = db.Exec("INSERT INTO users (username, password, SO, pet_id) VALUES (?, ?, NULL, NULL)", req.Username, hashedPassword)
	if err != nil {
		// Detect sqlite unique constraint on username and return a helpful error
		if strings.Contains(err.Error(), "UNIQUE constraint failed") && strings.Contains(err.Error(), "users.username") {
			logMessage("create_user_failed", map[string]interface{}{"username": req.Username, "reason": "username_taken"})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid_request", "reason": "username_taken"})
			return
		}
		logMessage("create_user_error", map[string]interface{}{"error": err.Error()})
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	logMessage("user_login", map[string]interface{}{"username": req.Username})

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created successfully"))
}

// createPetHandler handles the creation of a new pet for the authenticated user.
// Endpoint: POST /create_pet
// Request Body:
// - name: The name of the new pet.
// Response:
// - 201 Created on success.
// - 400 Bad Request if the request body is invalid.
// - 401 Unauthorized if the user is not authenticated.
// - 500 Internal Server Error if pet creation fails.
func createPetHandler(w http.ResponseWriter, r *http.Request) {
	token, err := oauth2Server.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userIDStr := token.GetUserID()
	if userIDStr == "" {
		http.Error(w, "Token must be issued with a user id (use password grant)", http.StatusUnauthorized)
		return
	}
	// Ensure userID is the DB primary key integer
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user id in token", http.StatusBadRequest)
		return
	}

	// No request body required for create_pet; the server will create a default pet for the caller.
	// Keep compatibility: attempt to decode but ignore any provided name.
	type CreatePetRequest struct {
		Name string `json:"name"`
	}

	var req CreatePetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure the user exists and then create the pet in a transaction so we can
	// set users.pet_id to the newly-created pet id (avoids foreign key issues).
	var existingUserID int
	if err := db.QueryRow("SELECT id FROM users WHERE id = ?", userIDInt).Scan(&existingUserID); err != nil {
		if err == sql.ErrNoRows {
			logMessage("create_pet_failed", map[string]interface{}{"user_id": userIDStr, "reason": "user_not_found"})
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		logMessage("create_pet_error", map[string]interface{}{"error": err.Error(), "user_id": userIDStr, "pet_name": req.Name})
		http.Error(w, "Error checking user", http.StatusInternalServerError)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		logMessage("create_pet_error", map[string]interface{}{"error": err.Error(), "user_id": userIDStr, "pet_name": req.Name})
		http.Error(w, "Error creating pet", http.StatusInternalServerError)
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	res, err := tx.Exec("INSERT INTO pets (main_owner, owner2, money, health, hunger, happiness) VALUES (?, NULL, 0, 100, 100, 100)", userIDInt)
	if err != nil {
		tx.Rollback()
		logMessage("create_pet_error", map[string]interface{}{"error": err.Error(), "user_id": userIDStr})
		http.Error(w, "Error creating pet", http.StatusInternalServerError)
		return
	}
	petID64, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		logMessage("create_pet_error", map[string]interface{}{"error": err.Error(), "user_id": userIDStr})
		http.Error(w, "Error creating pet", http.StatusInternalServerError)
		return
	}
	petID := int(petID64)

	// Update the user's pet_id to the new pet id.
	if _, err := tx.Exec("UPDATE users SET pet_id = ? WHERE id = ?", petID, userIDInt); err != nil {
		tx.Rollback()
		logMessage("create_pet_error", map[string]interface{}{"error": err.Error(), "user_id": userIDStr, "pet_id": petID})
		http.Error(w, "Error linking pet to user", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		logMessage("create_pet_error", map[string]interface{}{"error": err.Error(), "user_id": userIDStr, "pet_name": req.Name, "pet_id": petID})
		http.Error(w, "Error finalizing pet creation", http.StatusInternalServerError)
		return
	}

	logMessage("create_pet", map[string]interface{}{"user_id": userIDStr, "pet_id": petID})

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Pet created successfully"))
}

// validateCredentials checks the user's username and password and returns the user's DB id on success.
func validateCredentials(username, password string) (int, error) {
	var id int
	var hashedPassword string
	row := db.QueryRow("SELECT id, password FROM users WHERE username = ?", username)
	if err := row.Scan(&id, &hashedPassword); err != nil {
		logMessage("user_login_failed", map[string]interface{}{"username": username, "reason": "not_found"})
		return 0, fmt.Errorf("invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) != nil {
		logMessage("user_login_failed", map[string]interface{}{"username": username, "reason": "bad_password"})
		return 0, fmt.Errorf("invalid credentials")
	}
	logMessage("user_login_success", map[string]interface{}{"username": username, "user_id": id})
	return id, nil
}

// connectHandler validates username/password and then delegates to the OAuth2 token endpoint
// to obtain a token using the password grant. It ensures credentials are checked before
// returning a token and that the issued token is user-scoped (user id is in the token).
func connectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Expect JSON body with username and password
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate credentials first
	_, err := validateCredentials(creds.Username, creds.Password)
	if err != nil {
		logMessage("connect_failed", map[string]interface{}{"username": creds.Username, "reason": "invalid_credentials"})
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Build a form POST to the local /token handler to request a password grant token.
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", creds.Username)
	form.Set("password", creds.Password)
	form.Set("client_id", oauthClientID)
	form.Set("client_secret", oauthClientSecret)

	// Create a new request to the token endpoint handler, reusing the server's handler directly
	req, err := http.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use ResponseRecorder-like pattern: call the existing token handler directly
	rr := httptest.NewRecorder()
	oauth2Server.HandleTokenRequest(rr, req)

	// Log token issuance attempt result
	if rr.Code >= 200 && rr.Code < 300 {
		logMessage("connect_token_issued", map[string]interface{}{"username": creds.Username, "status": rr.Code})
	} else {
		logMessage("connect_token_failed", map[string]interface{}{"username": creds.Username, "status": rr.Code, "body": rr.Body.String()})
	}

	// Copy the response from the recorder to the real writer
	for k, vals := range rr.HeaderMap {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(rr.Code)
	w.Write(rr.Body.Bytes())
}

// addCoOwnerHandler handles adding another user as a co-owner of the caller's pet.
// Endpoint: POST /add_co_owner
// Request Body:
// - username: The username of the user to add as a co-owner.
// Response:
// - 200 OK on success.
// - 400 Bad Request if the request body is invalid.
// - 401 Unauthorized if the user is not authenticated.
// - 404 Not Found if user or pet not found.
// - 500 Internal Server Error if the update fails.
func addCoOwnerHandler(w http.ResponseWriter, r *http.Request) {
	token, err := oauth2Server.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userIDStr := token.GetUserID()
	if userIDStr == "" {
		http.Error(w, "Token must be user-scoped", http.StatusUnauthorized)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	type AddCoOwnerRequest struct {
		Username string `json:"username"`
	}

	var req AddCoOwnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find the current user's pet_id
	var petID sql.NullInt64
	if err := db.QueryRow("SELECT pet_id FROM users WHERE id = ?", userID).Scan(&petID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error reading user", http.StatusInternalServerError)
		return
	}
	if !petID.Valid {
		http.Error(w, "Caller has no pet to add a co-owner to", http.StatusBadRequest)
		return
	}

	// Resolve the target user's id by username
	var targetUserID int
	if err := db.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&targetUserID); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Target user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error reading target user", http.StatusInternalServerError)
		return
	}

	// Update pets.owner2 where appropriate
	res, err := db.Exec("UPDATE pets SET owner2 = ? WHERE id = ? AND owner2 IS NULL", targetUserID, int(petID.Int64))
	if err != nil {
		http.Error(w, "Error adding co-owner", http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Pet already has a co-owner or not found", http.StatusBadRequest)
		return
	}

	logMessage("add_co_owner", map[string]interface{}{"pet_id": petID.Int64, "new_owner_username": req.Username, "new_owner_id": targetUserID})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Co-owner added successfully"))
}

// Check if the OAuth2 server is running
func checkOAuthServer() error {
	resp, err := http.Get("http://localhost:8080/health") // Replace with actual OAuth2 server health endpoint
	if err != nil {
		return fmt.Errorf("failed to connect to OAuth2 server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OAuth2 server is not healthy, status code: %d", resp.StatusCode)
	}
	return nil
}

func init_servers() {
	// Load .env for local development so os.Getenv picks up values from the .env file.
	// Ignore error: absence of .env is okay in production.
	_ = godotenv.Load(".env")
	var err error
	db, err = sql.Open("sqlite3", "./game.db")
	if err != nil {
		logMessage("fatal", map[string]interface{}{"error": err.Error(), "context": "db_connect"})
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	var query []byte
	query, err = os.ReadFile("./schema.sql")

	if err != nil {
		logMessage("fatal", map[string]interface{}{"error": err.Error(), "context": "read_schema"})
		log.Fatalf("Failed to open schema file: %v", err)
	}

	// Enable foreign key enforcement for SQLite
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		logMessage("fatal", map[string]interface{}{"error": err.Error(), "context": "enable_foreign_keys"})
		log.Fatalf("Failed to enable SQLite foreign keys: %v", err)
	}

	// Execute the schema SQL. Use Exec to run CREATE TABLE statements and check for errors.
	if _, err := db.Exec(string(query)); err != nil {
		logMessage("fatal", map[string]interface{}{"error": err.Error(), "context": "exec_schema"})
		log.Fatalf("Failed to execute schema SQL: %v", err)
	}

	// Use the OAuth2 setup functions from oauth_setup.go
	// Pass clientID and clientSecret to the OAuth2 setup functions
	// Read client credentials from environment variables (fall back to defaults only for dev)
	clientID := os.Getenv("OAUTH2_CLIENT_ID")
	if clientID == "" {
		clientID = "motchi_app"
	}
	clientSecret := os.Getenv("OAUTH2_CLIENT_SECRET")
	if clientSecret == "" {
		clientSecret = "dev_secret_change_me"
	}
	// Store for visibility in main
	oauthClientID = clientID
	oauthClientSecret = clientSecret
	manager := initOAuth2Manager(clientID, clientSecret)
	oauth2Server = initOAuth2Server(manager)

	// Load environment variables for OAuth2 client credentials and log level
	logLevel = os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "production" // Default to production
	}
}

// main initializes the server and sets up the HTTP routes.
// Routes:
// - POST /create_user: Create a new user account.
// - POST /create_pet: Create a new pet for the authenticated user.
// - POST /add_co_owner: Add another user as a co-owner of a pet.
// - GET /ws: Establish a WebSocket connection.
func main() {
	init_servers()

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request's grant_type is one we allow. We only permit
		// the password grant and the refresh_token grant. This prevents
		// client_credentials or other grants from issuing client-only tokens.
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid_request", http.StatusBadRequest)
			return
		}
		grant := r.Form.Get("grant_type")
		if grant != "password" && grant != "refresh_token" {
			// Log and return an OAuth2-style error response
			logMessage("token_grant_denied", map[string]interface{}{"grant_type": grant})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"unsupported_grant_type"}`))
			return
		}
		oauth2Server.HandleTokenRequest(w, r)
	})
	// Simple token validation endpoint useful for manual testing
	http.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		_, err := oauth2Server.ValidationBearerToken(r)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Token valid"))
	})
	http.HandleFunc("/create_user", createUserHandler)
	http.HandleFunc("/create_pet", createPetHandler)
	http.HandleFunc("/add_co_owner", addCoOwnerHandler)
	http.HandleFunc("/connect", connectHandler)
	http.HandleFunc("/ws", websocketHandler)

	// Health endpoint so external checks (and our own check) succeed
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Log oauth client info (safe to show in development). Do not expose secrets in production logs.
	if logLevel == "development" {
		logMessage("oauth_client_info", map[string]interface{}{"client_id": oauthClientID, "client_secret": oauthClientSecret})
	} else {
		logMessage("oauth_client_info", map[string]interface{}{"client_id": oauthClientID})
	}
	logMessage("server_start", map[string]interface{}{"addr": ":8080"})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logMessage("server_failed", map[string]interface{}{"error": err.Error()})
		log.Fatalf("Server failed: %v", err)
	}
}

/*
Application Overview:
This application serves as the backend for a game where users can create accounts, manage pets, and interact in real-time.

Key Features:
- OAuth2 authentication for secure access.
- SQLite database for user and pet data storage.
- WebSocket support for real-time updates.
- REST APIs for user and pet management.

Environment Variables:
- OAUTH2_CLIENT_ID: The client ID for OAuth2 authentication.
- OAUTH2_CLIENT_SECRET: The client secret for OAuth2 authentication.
- LOG_LEVEL: The logging level ("development" or "production").

Endpoints:
1. POST /create_user:
   - Description: Create a new user account.
   - Request Body: {"email": "user@example.com", "password": "password123"}
   - Response: 201 Created on success.

2. POST /create_pet:
   - Description: Create a new pet for the authenticated user.
   - Request Body: {"name": "Fluffy"}
   - Response: 201 Created on success.

3. POST /add_co_owner:
   - Description: Add another user as a co-owner of a pet.
   - Request Body: {"pet_id": 1, "user_id": 2}
   - Response: 200 OK on success.

4. GET /ws:
   - Description: Establish a WebSocket connection for real-time communication.
   - Authentication: Requires a valid OAuth2 token.

Logging:
- Development: Logs all messages sent and received, and all requests.
- Production: Logs specific events (user login, pet creation, adding co-owners).
*/
