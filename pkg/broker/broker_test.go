package broker

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/keycloak-broker/pkg/keycloak"
	"github.com/labstack/echo/v4"
)

// newTestKeycloakClient creates a keycloak.Client backed by a mock server.
func newTestKeycloakClient(handler http.Handler) (*keycloak.Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	c := keycloak.NewTestClient(srv.URL, "dev-realm")
	return c, srv
}

func loadFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	return data
}

func mockKeycloakHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	clientsJSON := loadFixture(t, "_fixtures/get_client_response.json")
	discoveryJSON := loadFixture(t, "_fixtures/discovery_response.json")

	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Write(clientsJSON)
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
		case strings.HasPrefix(r.URL.Path, "/admin/realms/dev-realm/clients/") && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusNoContent)
		case strings.HasPrefix(r.URL.Path, "/admin/realms/dev-realm/clients/") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case r.URL.Path == "/realms/dev-realm/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write(discoveryJSON)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func TestGetCatalog(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v2/catalog", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	kc, srv := newTestKeycloakClient(mockKeycloakHandler(t))
	defer srv.Close()
	b := NewBroker(kc)

	if err := b.GetCatalog(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	json.Unmarshal(rec.Body.Bytes(), &body)
	services, ok := body["services"].([]any)
	if !ok || len(services) == 0 {
		t.Error("expected at least one service in catalog")
	}
}

func TestProvisionInstance_Success(t *testing.T) {
	e := echo.New()
	payload := `{
		"service_id": "fff5b36a-da19-4dc2-bd28-3dd331146290",
		"plan_id": "40627d0f-dedd-4d68-8111-2ebae510ba1b",
		"parameters": {
			"redirectURIs": ["https://myapp.example.com/callback"],
			"pkceEnabled": true
		}
	}`
	req := httptest.NewRequest(http.MethodPut, "/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("instance_id")
	c.SetParamValues("fe5556b9-8478-409b-ab2b-3c95ba06c5fc")

	// mock returns existing client, so provision should return 200 (idempotent)
	kc, srv := newTestKeycloakClient(mockKeycloakHandler(t))
	defer srv.Close()
	b := NewBroker(kc)

	if err := b.ProvisionInstance(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (already exists), got %d", rec.Code)
	}
	var body OSBClientResponse
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Parameters.ClientId != "fe5556b9-8478-409b-ab2b-3c95ba06c5fc" {
		t.Errorf("unexpected clientId: %s", body.Parameters.ClientId)
	}
}

func TestProvisionInstance_NewClient_AllParameters(t *testing.T) {
	discoveryJSON := loadFixture(t, "_fixtures/discovery_response.json")
	clientsJSON := loadFixture(t, "_fixtures/get_client_response.json")

	var capturedPayload keycloak.OIDCClientPayload
	getCalls := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodGet:
			getCalls++
			w.Header().Set("Content-Type", "application/json")
			if getCalls == 1 {
				w.Write([]byte("[]"))
			} else {
				w.Write(clientsJSON)
			}
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodPost:
			raw, _ := io.ReadAll(r.Body)
			json.Unmarshal(raw, &capturedPayload)
			w.WriteHeader(http.StatusCreated)
		case r.URL.Path == "/realms/dev-realm/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write(discoveryJSON)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}

	e := echo.New()
	payload := `{
		"service_id": "fff5b36a-da19-4dc2-bd28-3dd331146290",
		"plan_id": "40627d0f-dedd-4d68-8111-2ebae510ba1b",
		"parameters": {
			"redirectURIs": ["https://myapp.example.com/callback"],
			"standardFlowEnabled": false,
			"implicitFlowEnabled": true,
			"directAccessGrantsEnabled": true,
			"consentRequired": true,
			"serviceAccountsEnabled": true,
			"pkceEnabled": false,
			"refreshTokenLifetime": 600,
			"accessTokenLifetime": 300
		}
	}`
	req := httptest.NewRequest(http.MethodPut, "/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("instance_id")
	c.SetParamValues("fe5556b9-8478-409b-ab2b-3c95ba06c5fc")

	kc, srv := newTestKeycloakClient(http.HandlerFunc(handler))
	defer srv.Close()
	b := NewBroker(kc)

	if err := b.ProvisionInstance(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if capturedPayload.ClientId != "fe5556b9-8478-409b-ab2b-3c95ba06c5fc" {
		t.Errorf("provision: clientId = %q, want %q", capturedPayload.ClientId, "fe5556b9-8478-409b-ab2b-3c95ba06c5fc")
	}
	if capturedPayload.StandardFlowEnabled {
		t.Errorf("provision: standardFlowEnabled = %v, want false", capturedPayload.StandardFlowEnabled)
	}
	if !capturedPayload.ImplicitFlowEnabled {
		t.Errorf("provision: implicitFlowEnabled = %v, want true", capturedPayload.ImplicitFlowEnabled)
	}
	if !capturedPayload.DirectAccessGrantsEnabled {
		t.Errorf("provision: directAccessGrantsEnabled = %v, want true", capturedPayload.DirectAccessGrantsEnabled)
	}
	if !capturedPayload.ConsentRequired {
		t.Errorf("provision: consentRequired = %v, want true", capturedPayload.ConsentRequired)
	}
	if !capturedPayload.ServiceAccountsEnabled {
		t.Errorf("provision: serviceAccountsEnabled = %v, want true", capturedPayload.ServiceAccountsEnabled)
	}
	if _, ok := capturedPayload.Attributes["pkce.code.challenge.method"]; ok {
		t.Errorf("provision: pkce.code.challenge.method should not be set when pkceEnabled is false")
	}
	if v := capturedPayload.Attributes["client.session.max.lifespan"]; v != "600" {
		t.Errorf("provision: refreshTokenLifetime = %q, want %q", v, "600")
	}
	if v := capturedPayload.Attributes["access.token.lifespan"]; v != "300" {
		t.Errorf("provision: accessTokenLifetime = %q, want %q", v, "300")
	}
	if len(capturedPayload.RedirectURIs) != 1 || capturedPayload.RedirectURIs[0] != "https://myapp.example.com/callback" {
		t.Errorf("provision: redirectURIs = %v, want [https://myapp.example.com/callback]", capturedPayload.RedirectURIs)
	}
}

func TestUpdateInstance_AllParameters(t *testing.T) {
	discoveryJSON := loadFixture(t, "_fixtures/discovery_response.json")
	clientsJSON := loadFixture(t, "_fixtures/get_client_response.json")

	var capturedPayload keycloak.OIDCClientPayload
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Write(clientsJSON)
		case strings.HasPrefix(r.URL.Path, "/admin/realms/dev-realm/clients/") && r.Method == http.MethodPut:
			raw, _ := io.ReadAll(r.Body)
			json.Unmarshal(raw, &capturedPayload)
			w.WriteHeader(http.StatusNoContent)
		case r.URL.Path == "/realms/dev-realm/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			w.Write(discoveryJSON)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}

	e := echo.New()
	payload := `{
		"service_id": "fff5b36a-da19-4dc2-bd28-3dd331146290",
		"plan_id": "40627d0f-dedd-4d68-8111-2ebae510ba1b",
		"parameters": {
			"redirectURIs": ["https://new.example.com/callback"],
			"standardFlowEnabled": false,
			"implicitFlowEnabled": true,
			"directAccessGrantsEnabled": false,
			"consentRequired": true,
			"serviceAccountsEnabled": true,
			"pkceEnabled": false,
			"refreshTokenLifetime": 900,
			"accessTokenLifetime": 120
		}
	}`
	req := httptest.NewRequest(http.MethodPatch, "/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("instance_id")
	c.SetParamValues("fe5556b9-8478-409b-ab2b-3c95ba06c5fc")

	kc, srv := newTestKeycloakClient(http.HandlerFunc(handler))
	defer srv.Close()
	b := NewBroker(kc)

	if err := b.UpdateInstance(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if capturedPayload.StandardFlowEnabled {
		t.Errorf("update: standardFlowEnabled = %v, want false", capturedPayload.StandardFlowEnabled)
	}
	if !capturedPayload.ImplicitFlowEnabled {
		t.Errorf("update: implicitFlowEnabled = %v, want true", capturedPayload.ImplicitFlowEnabled)
	}
	if capturedPayload.DirectAccessGrantsEnabled {
		t.Errorf("update: directAccessGrantsEnabled = %v, want false", capturedPayload.DirectAccessGrantsEnabled)
	}
	if !capturedPayload.ConsentRequired {
		t.Errorf("update: consentRequired = %v, want true", capturedPayload.ConsentRequired)
	}
	if !capturedPayload.ServiceAccountsEnabled {
		t.Errorf("update: serviceAccountsEnabled = %v, want true", capturedPayload.ServiceAccountsEnabled)
	}
	if _, ok := capturedPayload.Attributes["pkce.code.challenge.method"]; ok {
		t.Errorf("update: pkce.code.challenge.method should not be set when pkceEnabled is false")
	}
	if v := capturedPayload.Attributes["client.session.max.lifespan"]; v != "900" {
		t.Errorf("update: refreshTokenLifetime = %q, want %q", v, "900")
	}
	if v := capturedPayload.Attributes["access.token.lifespan"]; v != "120" {
		t.Errorf("update: accessTokenLifetime = %q, want %q", v, "120")
	}
	if len(capturedPayload.RedirectURIs) != 1 || capturedPayload.RedirectURIs[0] != "https://new.example.com/callback" {
		t.Errorf("update: redirectURIs = %v, want [https://new.example.com/callback]", capturedPayload.RedirectURIs)
	}
}

func TestProvisionInstance_InvalidInstanceID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/v2/service_instances/bad-id", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("instance_id")
	c.SetParamValues("bad-id")

	kc, srv := newTestKeycloakClient(mockKeycloakHandler(t))
	defer srv.Close()
	b := NewBroker(kc)

	if err := b.ProvisionInstance(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestProvisionInstance_InvalidServiceID(t *testing.T) {
	e := echo.New()
	payload := `{
		"service_id": "00000000-0000-4000-a000-000000000000",
		"plan_id": "40627d0f-dedd-4d68-8111-2ebae510ba1b"
	}`
	req := httptest.NewRequest(http.MethodPut, "/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("instance_id")
	c.SetParamValues("fe5556b9-8478-409b-ab2b-3c95ba06c5fc")

	kc, srv := newTestKeycloakClient(mockKeycloakHandler(t))
	defer srv.Close()
	b := NewBroker(kc)

	if err := b.ProvisionInstance(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestGetInstance_Found(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("instance_id")
	c.SetParamValues("fe5556b9-8478-409b-ab2b-3c95ba06c5fc")

	kc, srv := newTestKeycloakClient(mockKeycloakHandler(t))
	defer srv.Close()
	b := NewBroker(kc)

	if err := b.GetInstance(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var body OSBClientResponse
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Parameters.PKCEEnabled != true {
		t.Error("expected pkceEnabled true in response")
	}
	if body.Parameters.RefreshTokenLifetime != 600 {
		t.Errorf("expected refreshTokenLifetime 600, got %d", body.Parameters.RefreshTokenLifetime)
	}
	if body.Parameters.AccessTokenLifetime != 300 {
		t.Errorf("expected accessTokenLifetime 300, got %d", body.Parameters.AccessTokenLifetime)
	}
	if body.Parameters.TokenEndpoint == "" {
		t.Error("expected tokenEndpoint to be populated")
	}
	if len(body.Parameters.RedirectURIs) != 1 || body.Parameters.RedirectURIs[0] != "https://myapp.example.com/callback" {
		t.Errorf("expected redirectURIs [https://myapp.example.com/callback], got %v", body.Parameters.RedirectURIs)
	}
}

func TestGetInstance_NotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v2/service_instances/a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("instance_id")
	c.SetParamValues("a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d")

	// return empty clients list
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/admin/realms/dev-realm/clients" && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
	kc, srv := newTestKeycloakClient(http.HandlerFunc(handler))
	defer srv.Close()
	b := NewBroker(kc)

	if err := b.GetInstance(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestDeprovisionInstance_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("instance_id")
	c.SetParamValues("fe5556b9-8478-409b-ab2b-3c95ba06c5fc")

	kc, srv := newTestKeycloakClient(mockKeycloakHandler(t))
	defer srv.Close()
	b := NewBroker(kc)

	if err := b.DeprovisionInstance(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestBindInstance_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc/service_bindings/db59931a-70a6-43c1-8885-b0c6b1c194d4", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("instance_id", "binding_id")
	c.SetParamValues("fe5556b9-8478-409b-ab2b-3c95ba06c5fc", "db59931a-70a6-43c1-8885-b0c6b1c194d4")

	kc, srv := newTestKeycloakClient(mockKeycloakHandler(t))
	defer srv.Close()
	b := NewBroker(kc)

	if err := b.BindInstance(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var body OSBBindingResponse
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Credentials.ClientId != "fe5556b9-8478-409b-ab2b-3c95ba06c5fc" {
		t.Errorf("unexpected clientId in binding: %s", body.Credentials.ClientId)
	}
	if body.Credentials.ClientSecret != "test-secret-value" {
		t.Errorf("unexpected clientSecret in binding: %s", body.Credentials.ClientSecret)
	}
	if body.Credentials.TokenEndpoint == "" {
		t.Error("expected token endpoint in binding credentials")
	}
	if len(body.Credentials.RedirectURIs) != 1 || body.Credentials.RedirectURIs[0] != "https://myapp.example.com/callback" {
		t.Errorf("expected redirectURIs [https://myapp.example.com/callback], got %v", body.Credentials.RedirectURIs)
	}
}

func TestUnbindInstance_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/v2/service_instances/fe5556b9-8478-409b-ab2b-3c95ba06c5fc/service_bindings/db59931a-70a6-43c1-8885-b0c6b1c194d4", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("instance_id", "binding_id")
	c.SetParamValues("fe5556b9-8478-409b-ab2b-3c95ba06c5fc", "db59931a-70a6-43c1-8885-b0c6b1c194d4")

	kc, srv := newTestKeycloakClient(mockKeycloakHandler(t))
	defer srv.Close()
	b := NewBroker(kc)

	if err := b.UnbindInstance(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestKeycloakClientToOSB_MapsAttributes(t *testing.T) {
	client := &keycloak.OIDCClientResponse{
		ClientId:              "test-client",
		Secret:                "test-secret",
		Protocol:              "openid-connect",
		StandardFlowEnabled:   true,
		ServiceAccountsEnabled: true,
		Issuer:                "https://kc.example.com/realms/test",
		TokenEndpoint:         "https://kc.example.com/realms/test/protocol/openid-connect/token",
		Attributes: map[string]string{
			"service_id":                   "svc-1",
			"plan_id":                      "plan-1",
			"pkce.code.challenge.method":   "S256",
			"client.session.max.lifespan":  "600",
			"access.token.lifespan":        "300",
		},
	}
	osb := keycloakClientToOSB(client)
	if osb.ServiceId != "svc-1" {
		t.Errorf("expected serviceId svc-1, got %s", osb.ServiceId)
	}
	if !osb.Parameters.PKCEEnabled {
		t.Error("expected pkceEnabled true")
	}
	if osb.Parameters.RefreshTokenLifetime != 600 {
		t.Errorf("expected refreshTokenLifetime 600, got %d", osb.Parameters.RefreshTokenLifetime)
	}
	if osb.Parameters.AccessTokenLifetime != 300 {
		t.Errorf("expected accessTokenLifetime 300, got %d", osb.Parameters.AccessTokenLifetime)
	}
	if !osb.Parameters.StandardFlowEnabled {
		t.Error("expected standardFlowEnabled true")
	}
	if osb.Parameters.TokenEndpoint != "https://kc.example.com/realms/test/protocol/openid-connect/token" {
		t.Errorf("expected tokenEndpoint, got %s", osb.Parameters.TokenEndpoint)
	}
	if osb.Parameters.Issuer != "https://kc.example.com/realms/test" {
		t.Errorf("expected issuer, got %s", osb.Parameters.Issuer)
	}
}

func TestKeycloakClientToOSBBinding_MapsCredentials(t *testing.T) {
	client := &keycloak.OIDCClientResponse{
		ClientId:      "test-client",
		Secret:        "test-secret",
		RedirectURIs:  []string{"https://example.com/cb"},
		Issuer:        "https://kc.example.com/realms/test",
		TokenEndpoint: "https://kc.example.com/realms/test/protocol/openid-connect/token",
		Attributes: map[string]string{
			"service_id": "svc-1",
			"plan_id":    "plan-1",
		},
	}
	binding := keycloakClientToOSBBinding(client)
	if binding.Credentials.ClientId != "test-client" {
		t.Errorf("expected clientId test-client, got %s", binding.Credentials.ClientId)
	}
	if binding.Credentials.ClientSecret != "test-secret" {
		t.Errorf("expected clientSecret test-secret, got %s", binding.Credentials.ClientSecret)
	}
	if binding.Metadata.ServiceId != "svc-1" {
		t.Errorf("expected serviceId svc-1, got %s", binding.Metadata.ServiceId)
	}
	if len(binding.Credentials.RedirectURIs) != 1 || binding.Credentials.RedirectURIs[0] != "https://example.com/cb" {
		t.Errorf("expected redirectURIs [https://example.com/cb], got %v", binding.Credentials.RedirectURIs)
	}
	if binding.Credentials.Issuer != "https://kc.example.com/realms/test" {
		t.Errorf("expected issuer, got %s", binding.Credentials.Issuer)
	}
	if binding.Credentials.TokenEndpoint != "https://kc.example.com/realms/test/protocol/openid-connect/token" {
		t.Errorf("expected tokenEndpoint, got %s", binding.Credentials.TokenEndpoint)
	}
}

// Suppress logger init noise
func init() {
	_ = time.Now()
}
