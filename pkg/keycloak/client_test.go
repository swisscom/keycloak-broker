package keycloak

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// newTestClient creates a Client pointing at the given test server.
func newTestClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    baseURL,
		realm:      "dev-realm",
		admin:      "admin",
		password:   "admin",
		token:      "mock-admin-token",
		expAt:      time.Now().Add(5 * time.Minute),
	}
}

func loadFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	return data
}

func TestGetClient_Found(t *testing.T) {
	clientsJSON := loadFixture(t, "_fixtures/get_client_response.json")
	discoveryJSON := loadFixture(t, "_fixtures/discovery_response.json")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Write(clientsJSON)
		case r.URL.Path == "/realms/dev-realm/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write(discoveryJSON)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	resp, err := c.GetClient(context.Background(), "fe5556b9-8478-409b-ab2b-3c95ba06c5fc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ClientId != "fe5556b9-8478-409b-ab2b-3c95ba06c5fc" {
		t.Errorf("expected clientId fe5556b9-8478-409b-ab2b-3c95ba06c5fc, got %s", resp.ClientId)
	}
	if resp.Secret != "test-secret-value" {
		t.Errorf("expected secret test-secret-value, got %s", resp.Secret)
	}
	if resp.Attributes["pkce.code.challenge.method"] != "S256" {
		t.Error("expected PKCE S256 attribute")
	}
	if resp.Attributes["client.session.max.lifespan"] != "600" {
		t.Error("expected refresh token lifetime 600")
	}
	if resp.Attributes["access.token.lifespan"] != "300" {
		t.Error("expected access token lifetime 300")
	}
	// discovery endpoints should be populated
	if resp.TokenEndpoint == "" {
		t.Error("expected token endpoint to be populated from discovery")
	}
	if resp.JWKSURI == "" {
		t.Error("expected JWKS URI to be populated from discovery")
	}
}

func TestGetClient_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/admin/realms/dev-realm/clients":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
		case r.URL.Path == "/realms/dev-realm/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{}"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.GetClient(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateClient_Success(t *testing.T) {
	clientsJSON := loadFixture(t, "_fixtures/get_client_response.json")
	discoveryJSON := loadFixture(t, "_fixtures/discovery_response.json")

	var capturedPayload OIDCClientPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &capturedPayload)
			w.WriteHeader(http.StatusCreated)
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Write(clientsJSON)
		case r.URL.Path == "/realms/dev-realm/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write(discoveryJSON)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	resp, err := c.CreateClient(context.Background(),
		"fe5556b9-8478-409b-ab2b-3c95ba06c5fc",
		"fff5b36a-da19-4dc2-bd28-3dd331146290",
		"40627d0f-dedd-4d68-8111-2ebae510ba1b",
		&OIDCClientParameters{
			RedirectURIs:              []string{"https://myapp.example.com/callback"},
			DirectAccessGrantsEnabled: true,
			RefreshTokenLifetime:      600,
			AccessTokenLifetime:       300,
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ClientId != "fe5556b9-8478-409b-ab2b-3c95ba06c5fc" {
		t.Errorf("unexpected clientId: %s", resp.ClientId)
	}
	// PKCE should default to true (nil PKCEEnabled)
	if capturedPayload.Attributes["pkce.code.challenge.method"] != "S256" {
		t.Error("expected PKCE S256 in create payload when PKCEEnabled is nil")
	}
	// StandardFlowEnabled should default to true
	if !capturedPayload.StandardFlowEnabled {
		t.Error("expected standardFlowEnabled to default to true")
	}
	if capturedPayload.Attributes["client.session.max.lifespan"] != "600" {
		t.Error("expected refresh token lifetime attribute")
	}
	if capturedPayload.Attributes["access.token.lifespan"] != "300" {
		t.Error("expected access token lifetime attribute")
	}
}

func TestCreateClient_PKCEDisabled(t *testing.T) {
	clientsJSON := loadFixture(t, "_fixtures/get_client_response.json")
	discoveryJSON := loadFixture(t, "_fixtures/discovery_response.json")

	var capturedPayload OIDCClientPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &capturedPayload)
			w.WriteHeader(http.StatusCreated)
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Write(clientsJSON)
		case r.URL.Path == "/realms/dev-realm/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write(discoveryJSON)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	pkceOff := false
	_, err := c.CreateClient(context.Background(),
		"fe5556b9-8478-409b-ab2b-3c95ba06c5fc",
		"fff5b36a-da19-4dc2-bd28-3dd331146290",
		"40627d0f-dedd-4d68-8111-2ebae510ba1b",
		&OIDCClientParameters{
			PKCEEnabled: &pkceOff,
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := capturedPayload.Attributes["pkce.code.challenge.method"]; ok {
		t.Error("expected no PKCE attribute when explicitly disabled")
	}
}

func TestCreateClient_StandardFlowDisabled(t *testing.T) {
	clientsJSON := loadFixture(t, "_fixtures/get_client_response.json")
	discoveryJSON := loadFixture(t, "_fixtures/discovery_response.json")

	var capturedPayload OIDCClientPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &capturedPayload)
			w.WriteHeader(http.StatusCreated)
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Write(clientsJSON)
		case r.URL.Path == "/realms/dev-realm/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write(discoveryJSON)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	stdOff := false
	_, err := c.CreateClient(context.Background(),
		"fe5556b9-8478-409b-ab2b-3c95ba06c5fc",
		"fff5b36a-da19-4dc2-bd28-3dd331146290",
		"40627d0f-dedd-4d68-8111-2ebae510ba1b",
		&OIDCClientParameters{
			StandardFlowEnabled:    &stdOff,
			ServiceAccountsEnabled: true,
		})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedPayload.StandardFlowEnabled {
		t.Error("expected standardFlowEnabled to be false")
	}
	if !capturedPayload.ServiceAccountsEnabled {
		t.Error("expected serviceAccountsEnabled to be true")
	}
}

func TestCreateClient_Conflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.CreateClient(context.Background(),
		"fe5556b9-8478-409b-ab2b-3c95ba06c5fc",
		"fff5b36a-da19-4dc2-bd28-3dd331146290",
		"40627d0f-dedd-4d68-8111-2ebae510ba1b",
		&OIDCClientParameters{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrConflict) {
		t.Errorf("expected ErrConflict, got %v", err)
	}
}

func TestDeleteClient_Success(t *testing.T) {
	clientsJSON := loadFixture(t, "_fixtures/get_client_response.json")
	discoveryJSON := loadFixture(t, "_fixtures/discovery_response.json")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Write(clientsJSON)
		case r.URL.Path == "/admin/realms/dev-realm/clients/internal-uuid-1234" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case r.URL.Path == "/realms/dev-realm/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write(discoveryJSON)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	err := c.DeleteClient(context.Background(), "fe5556b9-8478-409b-ab2b-3c95ba06c5fc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteClient_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	err := c.DeleteClient(context.Background(), "nonexistent-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateClient_MergesFields(t *testing.T) {
	clientsJSON := loadFixture(t, "_fixtures/get_client_response.json")
	discoveryJSON := loadFixture(t, "_fixtures/discovery_response.json")

	var capturedPayload OIDCClientPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Write(clientsJSON)
		case r.URL.Path == "/admin/realms/dev-realm/clients/internal-uuid-1234" && r.Method == http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &capturedPayload)
			w.WriteHeader(http.StatusNoContent)
		case r.URL.Path == "/realms/dev-realm/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write(discoveryJSON)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	implicitOn := true
	refreshTTL := 900
	_, err := c.UpdateClient(context.Background(), "fe5556b9-8478-409b-ab2b-3c95ba06c5fc", &OIDCClientUpdatePayload{
		ImplicitFlowEnabled:  &implicitOn,
		RefreshTokenLifetime: &refreshTTL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// updated field
	if !capturedPayload.ImplicitFlowEnabled {
		t.Error("expected implicitFlowEnabled to be true")
	}
	if capturedPayload.Attributes["client.session.max.lifespan"] != "900" {
		t.Errorf("expected refresh token lifetime 900, got %s", capturedPayload.Attributes["client.session.max.lifespan"])
	}
	// preserved fields (not in update payload)
	if !capturedPayload.StandardFlowEnabled {
		t.Error("expected standardFlowEnabled to be preserved as true")
	}
	if !capturedPayload.DirectAccessGrantsEnabled {
		t.Error("expected directAccessGrantsEnabled to be preserved as true")
	}
	if capturedPayload.Attributes["access.token.lifespan"] != "300" {
		t.Error("expected access token lifetime to be preserved as 300")
	}
}

func TestGetToken_Success(t *testing.T) {
	tokenJSON := loadFixture(t, "_fixtures/token_response.json")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/realms/master/protocol/openid-connect/token" {
			w.Header().Set("Content-Type", "application/json")
			w.Write(tokenJSON)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    srv.URL,
		realm:      "dev-realm",
		admin:      "admin",
		password:   "admin",
	}
	token, err := c.getToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "mock-admin-token" {
		t.Errorf("expected mock-admin-token, got %s", token)
	}
}

func TestGetToken_Cached(t *testing.T) {
	calls := 0
	tokenJSON := loadFixture(t, "_fixtures/token_response.json")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.Write(tokenJSON)
	}))
	defer srv.Close()

	c := &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    srv.URL,
		realm:      "dev-realm",
		admin:      "admin",
		password:   "admin",
	}
	c.getToken(context.Background())
	c.getToken(context.Background())
	if calls != 1 {
		t.Errorf("expected 1 token request (cached), got %d", calls)
	}
}

func TestHealthCheck_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/realms/dev-realm" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"realm":"dev-realm"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	err := c.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHealthCheck_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("unavailable"))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	err := c.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
