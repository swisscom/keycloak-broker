package keycloak

// OIDCClientParameters represents Keycloak OIDC client parameters
type OIDCClientParameters struct {
	RedirectURIs        []string          `json:"redirectUris"`
	PublicClient        bool              `json:"publicClient"`
	ConsentRequired     bool              `json:"consentRequired"`
	ImplicitFlowEnabled bool              `json:"implicitFlowEnabled"`
	Attributes          map[string]string `json:"attributes,omitempty"`
}

// OIDCClientPayload represents a Keycloak OIDC client payload for the admin API.
type OIDCClientPayload struct {
	ClientId                  string            `json:"clientId"`
	Name                      string            `json:"name"`
	Description               string            `json:"description"`
	Enabled                   bool              `json:"enabled"`
	Protocol                  string            `json:"protocol"`
	PublicClient              bool              `json:"publicClient"`
	RedirectURIs              []string          `json:"redirectUris"`
	ConsentRequired           bool              `json:"consentRequired"`
	StandardFlowEnabled       bool              `json:"standardFlowEnabled"`
	ImplicitFlowEnabled       bool              `json:"implicitFlowEnabled"`
	DirectAccessGrantsEnabled bool              `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool              `json:"serviceAccountsEnabled"`
	Attributes                map[string]string `json:"attributes,omitempty"`
}

// OIDCClientResponse represents a Keycloak OIDC client response.
type OIDCClientResponse struct {
	Id                        string            `json:"id"` // internal UUID
	ClientId                  string            `json:"clientId"`
	Name                      string            `json:"name"`
	Description               string            `json:"description"`
	SurrogateAuthRequired     bool              `json:"surrogateAuthRequired"`
	Enabled                   bool              `json:"enabled"`
	AlwaysDisplayInConsole    bool              `json:"alwaysDisplayInConsole"`
	ClientAuthenticatorType   string            `json:"clientAuthenticatorType"`
	Secret                    string            `json:"secret"`
	Protocol                  string            `json:"protocol"`
	PublicClient              bool              `json:"publicClient"`
	RedirectURIs              []string          `json:"redirectUris"`
	WebOrigins                []string          `json:"webOrigins"`
	ConsentRequired           bool              `json:"consentRequired"`
	StandardFlowEnabled       bool              `json:"standardFlowEnabled"`
	ImplicitFlowEnabled       bool              `json:"implicitFlowEnabled"`
	DirectAccessGrantsEnabled bool              `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool              `json:"serviceAccountsEnabled"`
	Issuer                    string            `json:"issuer"`
	DiscoveryEndpoint         string            `json:"discoveryEndpoint"`
	AuthorizationEndpoint     string            `json:"authorizationEndpoint"`
	TokenEndpoint             string            `json:"tokenEndpoint"`
	IntrospectionEndpoint     string            `json:"introspectionEndpoint"`
	UserInfoEndpoint          string            `json:"userInfoEndpoint"`
	EndSessionEndpoint        string            `json:"endSessionEndpoint"`
	JWKSURI                   string            `json:"jwksURI"`
	Attributes                map[string]string `json:"attributes"`
}

// OIDCDiscoveryResponse represents a Keycloak OIDC discovery endpoint response.
type OIDCDiscoveryResponse struct {
	Issuer                    string            `json:"issuer"`
	AuthorizationEndpoint     string            `json:"authorization_endpoint"`
	TokenEndpoint             string            `json:"token_endpoint"`
	IntrospectionEndpoint     string            `json:"introspection_endpoint"`
	UserInfoEndpoint          string            `json:"userinfo_endpoint"`
	EndSessionEndpoint        string            `json:"end_session_endpoint"`
	JWKSURI                   string            `json:"jwks_uri"`
}
