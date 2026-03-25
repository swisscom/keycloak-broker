package keycloak

// OIDCClientParameters represents Keycloak OIDC client parameters
type OIDCClientParameters struct {
	PublicClient              bool     `json:"public_client"`
	RedirectURIs              []string `json:"redirect_uris"`
	ImplicitFlowEnabled       bool     `json:"implicit_flow_enabled"`
}

// OIDCClientPayload represents a Keycloak OIDC client payload for the admin API.
type OIDCClientPayload struct {
	ClientID                  string   `json:"clientId"`
	Name                      string   `json:"name"`
	Description               string   `json:"description"`
	Enabled                   bool     `json:"enabled"`
	Protocol                  string   `json:"protocol"`
	PublicClient              bool     `json:"publicClient"`
	RedirectURIs              []string `json:"redirectUris"`
	StandardFlowEnabled       bool     `json:"standardFlowEnabled"`
	ImplicitFlowEnabled       bool     `json:"implicitFlowEnabled"`
	DirectAccessGrantsEnabled bool     `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool     `json:"serviceAccountsEnabled"`
}

// OIDCClientResponse represents a Keycloak OIDC client response.
type OIDCClientResponse struct {
	ID                        string   `json:"id,omitempty"` // internal UUID
	ClientID                  string   `json:"clientId,omitempty"`
	Name                      string   `json:"name,omitempty"`
	Description               string   `json:"description,omitempty"`
	SurrogateAuthRequired     bool     `json:"surrogateAuthRequired"`
	Enabled                   bool     `json:"enabled"`
	AlwaysDisplayInConsole    bool     `json:"alwaysDisplayInConsole"`
	ClientAuthenticatorType   string   `json:"clientAuthenticatorType,omitempty"`
	Secret                    string   `json:"secret,omitempty"`
	Protocol                  string   `json:"protocol,omitempty"`
	PublicClient              bool     `json:"publicClient"`
	RedirectURIs              []string `json:"redirectUris,omitempty"`
	WebOrigins                []string `json:"webOrigins,omitempty"`
	StandardFlowEnabled       bool     `json:"standardFlowEnabled"`
	ImplicitFlowEnabled       bool     `json:"implicitFlowEnabled"`
	DirectAccessGrantsEnabled bool     `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled    bool     `json:"serviceAccountsEnabled"`
}

/*
[{"id":"a3f297ce-c7f5-4ed9-8e79-165d5d359d6d","clientId":"fe5556b9-8478-409b-ab2b-3c95ba06c5fc","name":"fe5556b9-8478-409b-ab2b-3c95ba06c5fc","description":"managed OIDC client","surrogateAuthRequired":false,"enabled":true,"alwaysDisplayInConsole":false,"clientAuthenticatorType":"client-secret","secret":"YNqtYBjOoGi2MTg5JlbJBPLENLWd12KB","redirectUris":[],"webOrigins":[],"notBefore":0,"bearerOnly":false,"consentRequired":false,"standardFlowEnabled":true,"implicitFlowEnabled":false,"directAccessGrantsEnabled":false,"serviceAccountsEnabled":false,"publicClient":false,"frontchannelLogout":false,"protocol":"openid-connect","attributes":{"realm_client":"false","client.secret.creation.time":"1774463924","backchannel.logout.session.required":"true","backchannel.logout.revoke.offline.tokens":"false"},"authenticationFlowBindingOverrides":{},"fullScopeAllowed":true,"nodeReRegistrationTimeout":-1,"defaultClientScopes":["web-origins","acr","roles","profile","basic","email"],"optionalClientScopes":["address","phone","organization","offline_access","microprofile-jwt"],"access":{"view":true,"configure":true,"manage":true}}]
*/
