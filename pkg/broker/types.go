package broker

// OSBClientResponse represents an OSB API response of a client service instance.
type OSBClientResponse struct {
	ServiceId  string                      `json:"serviceId"`
	PlanId     string                      `json:"planId"`
	Parameters OSBClientResponseParameters `json:"parameters"`
}
type OSBClientResponseParameters struct {
	ClientId                  string            `json:"clientId"`
	ClientSecret              string            `json:"clientSecret"`
	Description               string            `json:"description"`
	Protocol                  string            `json:"protocol"`
	PublicClient              bool              `json:"publicClient,omitempty"`
	ClientAuthenticatorType   string            `json:"clientAuthenticatorType,omitempty"`
	RedirectURIs              []string          `json:"redirectURIs"`
	WebOrigins                []string          `json:"webOrigins,omitempty"`
	ConsentRequired           bool              `json:"consentRequired,omitempty"`
	StandardFlowEnabled       bool              `json:"standardFlowEnabled,omitempty"`
	ImplicitFlowEnabled       bool              `json:"implicitFlowEnabled,omitempty"`
	DirectAccessGrantsEnabled bool              `json:"directAccessGrantsEnabled,omitempty"`
	ServiceAccountsEnabled    bool              `json:"serviceAccountsEnabled,omitempty"`
	PKCEEnabled               bool              `json:"pkceEnabled,omitempty"`
	RefreshTokenLifespan      int               `json:"refreshTokenLifespan,omitempty"`
	Issuer                    string            `json:"issuer"`
	DiscoveryEndpoint         string            `json:"discoveryEndpoint"`
	AuthorizationEndpoint     string            `json:"authorizationEndpoint"`
	TokenEndpoint             string            `json:"tokenEndpoint"`
	IntrospectionEndpoint     string            `json:"introspectionEndpoint"`
	UserInfoEndpoint          string            `json:"userInfoEndpoint"`
	EndSessionEndpoint        string            `json:"endSessionEndpoint"`
	JWKSURI                   string            `json:"jwksURI"`
	Attributes                map[string]string `json:"attributes,omitempty"`
}

// OSBBindingResponse represents an OSB API response of a service instance binding.
type OSBBindingResponse struct {
	Metadata    OSBBindingResponseMetadata    `json:"metadata"`
	Credentials OSBBindingResponseCredentials `json:"credentials"`
}
type OSBBindingResponseMetadata struct {
	ServiceId string `json:"serviceId"`
	PlanId    string `json:"planId"`
}
type OSBBindingResponseCredentials struct {
	ClientId              string   `json:"clientId"`
	ClientSecret          string   `json:"clientSecret"`
	RedirectURIs          []string `json:"redirectURIs"`
	Issuer                string   `json:"issuer"`
	DiscoveryEndpoint     string   `json:"discoveryEndpoint"`
	AuthorizationEndpoint string   `json:"authorizationEndpoint"`
	TokenEndpoint         string   `json:"tokenEndpoint"`
	IntrospectionEndpoint string   `json:"introspectionEndpoint"`
	UserInfoEndpoint      string   `json:"userInfoEndpoint"`
	EndSessionEndpoint    string   `json:"endSessionEndpoint"`
	JWKSURI               string   `json:"jwksURI"`
}
