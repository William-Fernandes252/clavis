package server

import "github.com/William-Fernandes252/clavis/internal/store"

// Interface for a key-value store server.
type Server interface {
	// Start the server and listen for incoming requests
	Start(callback ...func()) error

	// Stop the server gracefully
	Shutdown()

	// Get the underlying store instance
	GetStore() (store.Store, error)
}
