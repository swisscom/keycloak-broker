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
	Attributes                map[string]string `json:"attributes"`
}

/*
[{"id":"a3f297ce-c7f5-4ed9-8e79-165d5d359d6d","clientId":"fe5556b9-8478-409b-ab2b-3c95ba06c5fc","name":"fe5556b9-8478-409b-ab2b-3c95ba06c5fc","description":"managed OIDC client","surrogateAuthRequired":false,"enabled":true,"alwaysDisplayInConsole":false,"clientAuthenticatorType":"client-secret","secret":"YNqtYBjOoGi2MTg5JlbJBPLENLWd12KB","redirectUris":[],"webOrigins":[],"notBefore":0,"bearerOnly":false,"consentRequired":false,"standardFlowEnabled":true,"implicitFlowEnabled":false,"directAccessGrantsEnabled":false,"serviceAccountsEnabled":false,"publicClient":false,"frontchannelLogout":false,"protocol":"openid-connect","attributes":{"realm_client":"false","client.secret.creation.time":"1774463924","backchannel.logout.session.required":"true","backchannel.logout.revoke.offline.tokens":"false"},"authenticationFlowBindingOverrides":{},"fullScopeAllowed":true,"nodeReRegistrationTimeout":-1,"defaultClientScopes":["web-origins","acr","roles","profile","basic","email"],"optionalClientScopes":["address","phone","organization","offline_access","microprofile-jwt"],"access":{"view":true,"configure":true,"manage":true}}]
*/
