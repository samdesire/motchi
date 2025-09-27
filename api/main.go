// Package main implements a backend server for a game application.
// It provides OAuth2-based authentication, WebSocket support for real-time communication,
// and REST APIs for user and pet management.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PetMoneyUpdate struct {
	PetID  int `json:"pet_id"`
	Amount int `json:"amount"`
}

var (
	db           *sql.DB
	oauth2Server *server.Server
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for simplicity
		},
	}
	connections   = make(map[int]*websocket.Conn) // Map of user ID to WebSocket connection
	connectionsMu sync.Mutex                      // Mutex to protect the connections map
	logLevel      string
)

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./game.db")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Use the OAuth2 setup functions from oauth_setup.go
	// Pass clientID and clientSecret to the OAuth2 setup functions
	clientID := os.Getenv("OAUTH2_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH2_CLIENT_SECRET")
	manager := initOAuth2Manager(clientID, clientSecret)
	oauth2Server = initOAuth2Server(manager)

	// Load environment variables for OAuth2 client credentials and log level
	logLevel = os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "production" // Default to production
	}
}

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
				log.Printf("Ping failed for user %d: %v", userID, err)
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
		log.Printf("Failed to upgrade connection: %v", err)
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
		log.Printf("Pong received from user %d", userID)
		return nil
	})

	go sendPingMessages(userID)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		// Process the message
		log.Printf("Message received: %s", string(message))

		// Notify other owner if applicable
		var updateData PetMoneyUpdate
		if err := json.Unmarshal(message, &updateData); err == nil {
			valid, newMoney, err := validateAndUpdatePetMoney(updateData.PetID, updateData.Amount)
			if err != nil {
				log.Printf("Error validating/updating pet money: %v", err)
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
					err = otherConn.WriteMessage(websocket.TextMessage, message)
					if err != nil {
						log.Printf("Error sending message to other owner: %v", err)
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
// - email: The email address of the new user.
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

	// Updated CreateUserRequest to include name
	type CreateUserRequest struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
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
	_, err = db.Exec("INSERT INTO users (name, email, password, SO, pet_id) VALUES (?, ?, ?, NULL, NULL)", req.Name, req.Email, hashedPassword)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	logMessage("user_login", map[string]interface{}{"email": req.Email})

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

	userID := token.GetUserID()

	type CreatePetRequest struct {
		Name string `json:"name"`
	}

	var req CreatePetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO pets (name, main_owner, money, health, hunger, happiness) VALUES (?, ?, 0, 100, 100, 100)", req.Name, userID)
	if err != nil {
		http.Error(w, "Error creating pet", http.StatusInternalServerError)
		return
	}

	logMessage("create_pet", map[string]interface{}{"user_id": userID, "pet_name": req.Name})

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Pet created successfully"))
}

// addCoOwnerHandler handles adding another user as a co-owner of a pet.
// Endpoint: POST /add_co_owner
// Request Body:
// - pet_id: The ID of the pet.
// - user_id: The ID of the user to add as a co-owner.
// Response:
// - 200 OK on success.
// - 400 Bad Request if the request body is invalid.
// - 401 Unauthorized if the user is not authenticated.
// - 500 Internal Server Error if the update fails.
func addCoOwnerHandler(w http.ResponseWriter, r *http.Request) {
	_, err := oauth2Server.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	type AddCoOwnerRequest struct {
		PetID  int `json:"pet_id"`
		UserID int `json:"user_id"`
	}

	var req AddCoOwnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE pets SET owner2 = ? WHERE id = ? AND owner2 IS NULL", req.UserID, req.PetID)
	if err != nil {
		http.Error(w, "Error adding co-owner", http.StatusInternalServerError)
		return
	}

	logMessage("add_co_owner", map[string]interface{}{"pet_id": req.PetID, "new_owner_id": req.UserID})

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

// main initializes the server and sets up the HTTP routes.
// Routes:
// - POST /create_user: Create a new user account.
// - POST /create_pet: Create a new pet for the authenticated user.
// - POST /add_co_owner: Add another user as a co-owner of a pet.
// - GET /ws: Establish a WebSocket connection.
func main() {
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		oauth2Server.HandleTokenRequest(w, r)
	})
	http.HandleFunc("/create_user", createUserHandler)
	http.HandleFunc("/create_pet", createPetHandler)
	http.HandleFunc("/add_co_owner", addCoOwnerHandler)
	http.HandleFunc("/ws", websocketHandler)

	// Call the checkOAuthServer function before starting the application
	if err := checkOAuthServer(); err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
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
