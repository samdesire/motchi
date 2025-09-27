package main

import (
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
)

// initOAuth2Manager initializes the OAuth2 manager.
func initOAuth2Manager(clientID, clientSecret string) *manage.Manager {
	manager := manage.NewDefaultManager()

	// Set up token generation
	manager.MapAccessGenerate(generates.NewAccessGenerate())

	// Set up client storage
	clientStore := store.NewClientStore()
	clientStore.Set(clientID, &models.Client{
		ID:     clientID,
		Secret: clientSecret,
		Domain: "http://localhost",
	})
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

	return oauth2Server
}
