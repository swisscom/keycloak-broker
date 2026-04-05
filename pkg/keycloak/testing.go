package keycloak

import (
	"net/http"
	"time"
)

// NewTestClient creates a Client for testing with a custom base URL and pre-set token.
func NewTestClient(baseURL, realm string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    baseURL,
		realm:      realm,
		admin:      "admin",
		password:   "admin",
		token:      "mock-admin-token",
		expAt:      time.Now().Add(5 * time.Minute),
	}
}
