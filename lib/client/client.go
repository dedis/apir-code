package client

// Client represents the client instance in both the IT and C models
type Client interface {
	Query()
	Reconstruct()
}

// General containts the elements needed by the clients of all schemes
type General struct {
	DBLength int
}

// return true if the query inputs are invalid
func invalidQueryInputs(index, numServers int) bool {
	return index < 0 && numServers < 1
}
