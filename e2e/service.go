package e2e

// ServiceStats holds the response type balance-service /info entry point.
type ServiceStats struct {
	RequestCount int `json:"request_count"`
}
