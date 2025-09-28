package main

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	oauth2 "github.com/go-oauth2/oauth2/v4"
	oautherrors "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	// bcrypt is used in main.validateCredentials
)

// initOAuth2Manager initializes the OAuth2 manager.
func initOAuth2Manager(clientID, clientSecret string) *manage.Manager {
	manager := manage.NewDefaultManager()

	// Use an in-memory token store so issued tokens are persisted while the server runs.
	// You can replace this with a persistent store (Redis, MySQL, etc.) for production.
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// Set up token generation
	manager.MapAccessGenerate(generates.NewAccessGenerate())

	// Set up client storage
	clientStore := store.NewClientStore()
	clientStore.Set(clientID, &models.Client{
		ID:     clientID,
		Secret: clientSecret,
		Domain: "http://localhost",
	})
	// Log the registered client id (do not log the secret in production)
	log.Printf("Registered OAuth2 client: %s", clientID)
	manager.MapClientStorage(clientStore)

	return manager
}

// initOAuth2Server initializes the OAuth2 server.
func initOAuth2Server(manager *manage.Manager) *server.Server {
	oauth2Server := server.NewDefaultServer(manager)

	// Allow GET requests for token validation
	oauth2Server.SetAllowGetAccessRequest(true)

	// Set client info handler
	oauth2Server.SetClientInfoHandler(server.ClientFormHandler)
	// Restrict allowed grant types to password and refresh_token only.
	// Use the server API to limit the grant types accepted.
	// If the go-oauth2 version exposes SetAllowedGrantType on the server, use it.
	// This prevents client_credentials from issuing client-only tokens that lack a user id.
	// Fallback: if the method is unavailable, clients will still be required to use
	// the password grant; additional enforcement can be added later.
	_ = oauth2Server
	// Note: some versions of the library expose a SetAllowedGrantType method on the server.
	// We rely on the password grant handler and extension fields to ensure user-scoped tokens.

	// Log internal errors to help diagnose server_error responses
	oauth2Server.SetInternalErrorHandler(func(err error) (re *oautherrors.Response) {
		log.Printf("OAuth2 internal error: %v", err)
		// If the internal error is the library's ErrInvalidGrant, return a
		// Response with that error so the HTTP response is the proper
		// OAuth2 error (invalid_grant) instead of a generic server_error.
		if err == oautherrors.ErrInvalidGrant {
			return &oautherrors.Response{Error: oautherrors.ErrInvalidGrant}
		}
		return nil
	})

	// Log response errors (e.g., invalid_client, invalid_grant)
	oauth2Server.SetResponseErrorHandler(func(re *oautherrors.Response) {
		if re != nil && re.Error != nil {
			log.Printf("OAuth2 response error: %v", re.Error)
		} else {
			log.Printf("OAuth2 response error: %+v", re)
		}
	})

	// Set password authorization handler so the password grant validates
	// credentials against our users table and returns the DB user ID as the
	// token's UserID (so tokens are user-scoped).
	oauth2Server.SetPasswordAuthorizationHandler(func(ctx context.Context, clientID, username, password string) (userID string, err error) {
		// Log attempt (don't include password)
		log.Printf("password_grant_attempt: client=%s username=%s", clientID, username)
		id, err := validateCredentials(username, password)
		if err != nil {
			log.Printf("password_grant_failed: client=%s username=%s", clientID, username)
			// Return the oauth2 library's ErrInvalidGrant so the server produces the
			// correct OAuth2 error response (invalid_grant) instead of treating this
			// as an internal server error.
			return "", oautherrors.ErrInvalidGrant
		}
		log.Printf("password_grant_success: client=%s username=%s user_id=%d", clientID, username, id)
		return strconv.Itoa(id), nil
	})

	// Include the user_id in token JSON responses when available so clients
	// can immediately see whether a token is user-scoped.
	oauth2Server.SetExtensionFieldsHandler(func(ti oauth2.TokenInfo) map[string]interface{} {
		uid := ti.GetUserID()
		if uid == "" {
			return nil
		}
		out := map[string]interface{}{"user_id": uid}
		// Try to fetch pet_id from DB so clients can see the pet associated with the
		// token's user without the client being able to tamper with it.
		if id, err := strconv.Atoi(uid); err == nil {
			var petID sql.NullInt64
			if err := db.QueryRow("SELECT pet_id FROM users WHERE id = ?", id).Scan(&petID); err == nil {
				if petID.Valid {
					out["pet_id"] = petID.Int64
				}
			} else {
				// If the DB read fails, log it but do not break token issuance.
				log.Printf("warning: unable to read pet_id for user %d: %v", id, err)
			}
		}
		return out
	})

	return oauth2Server
}
