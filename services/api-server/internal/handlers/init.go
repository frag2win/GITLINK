package handlers

import (
	"github.com/localrepo/api-server/internal/database"
	"github.com/localrepo/api-server/internal/socket"
)

var (
	gitClient *socket.GitClient
	db        *database.DB
)

// Init injects dependencies needed by the handlers that bypass services.
func Init(gc *socket.GitClient, database *database.DB) {
	gitClient = gc
	db = database
}
