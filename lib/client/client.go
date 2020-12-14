package client

// Client represents the client instance in both the IT and C models
type Client interface {
	Query()
	Reconstruct()
}
