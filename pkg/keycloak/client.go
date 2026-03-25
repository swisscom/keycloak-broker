package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/keycloak-broker/pkg/config"
	"github.com/keycloak-broker/pkg/logger"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	realm      string
	admin      string
	password   string

	mu    sync.Mutex
	token string
	expAt time.Time
}

// OIDCClient represents a Keycloak OIDC client returned by the admin API.
type OIDCClient struct {
	ID                        string   `json:"id,omitempty"` // internal UUID
	ClientID                  string   `json:"clientId"`
	Name                      string   `json:"name,omitempty"`
	Description               string   `json:"description,omitempty"`
	Enabled                   bool     `json:"enabled"`
	Protocol                  string   `json:"protocol,omitempty"`
	PublicClient              bool     `json:"publicClient"`
	RedirectURIs              []string `json:"redirectUris,omitempty"`
	StandardFlowEnabled       bool     `json:"standardFlowEnabled"`
	DirectAccessGrantsEnabled bool     `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool     `json:"serviceAccountsEnabled"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

var ErrNotFound = fmt.Errorf("not found")

func NewClient() *Client {
	cfg := config.Get()
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    strings.TrimRight(cfg.KeycloakURL, "/"),
		realm:      cfg.KeycloakRealm,
		admin:      cfg.KeycloakAdmin,
		password:   cfg.KeycloakPassword,
	}
}

// getToken returns a valid admin access token, refreshing if expired.
func (c *Client) getToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.expAt) {
		return c.token, nil
	}

	data := url.Values{
		"grant_type": {"password"},
		"client_id":  {"admin-cli"},
		"username":   {c.admin},
		"password":   {c.password},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/realms/master/protocol/openid-connect/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed (%d): %s", resp.StatusCode, body)
	}

	var tok tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", fmt.Errorf("decode token: %w", err)
	}

	c.token = tok.AccessToken
	c.expAt = time.Now().Add(time.Duration(tok.ExpiresIn-15) * time.Second) // 15s buffer
	return c.token, nil
}

// CreateClient creates an OIDC client in Keycloak and returns the instance_id.
func (c *Client) CreateClient(ctx context.Context, instanceId string, public bool) (string, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return "", err
	}

	oidc := OIDCClient{
		ClientID:                  instanceId,
		Name:                      instanceId,
		Description:               "managed OIDC client",
		Enabled:                   true,
		Protocol:                  "openid-connect",
		PublicClient:              public,
		StandardFlowEnabled:       true,
		DirectAccessGrantsEnabled: false,
		ServiceAccountsEnabled:    false,
	}
	body, err := json.Marshal(oidc)
	if err != nil {
		return "", fmt.Errorf("marshal client: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/admin/realms/%s/clients", c.baseURL, c.realm),
		bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create client: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return "", fmt.Errorf("client with instance_id [%s] already exists", instanceId)
	}
	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create client failed (%d): %s", resp.StatusCode, respBody)
	}

	logger.Info("created keycloak client with instance_id [%s]", instanceId)
	return instanceId, nil
}

// GetClient looks up an OIDC client by instanceId and returns it.
func (c *Client) GetClient(ctx context.Context, instanceId string) (*OIDCClient, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/admin/realms/%s/clients?clientId=%s", c.baseURL, c.realm, url.QueryEscape(instanceId)),
		nil)
	if err != nil {
		return nil, fmt.Errorf("build get request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get client: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get client failed (%d): %s", resp.StatusCode, respBody)
	}

	var clients []OIDCClient
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return nil, fmt.Errorf("decode clients: %w", err)
	}

	for _, cl := range clients {
		if cl.ClientID == instanceId {
			return &cl, nil
		}
	}
	return nil, fmt.Errorf("client instance_id [%s] %w", instanceId, ErrNotFound)
}

// DeleteClient removes an OIDC client by clientId.
func (c *Client) DeleteClient(ctx context.Context, instanceId string) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return err
	}

	// we need to the internal UUID to delete the client
	oidc, err := c.GetClient(ctx, instanceId)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/admin/realms/%s/clients/%s", c.baseURL, c.realm, oidc.ID),
		nil)
	if err != nil {
		return fmt.Errorf("build delete request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete client: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		logger.Info("deleted keycloak client instance_id [%s]", instanceId)
		return nil
	}
	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("delete client failed (%d): %s", resp.StatusCode, respBody)
}
