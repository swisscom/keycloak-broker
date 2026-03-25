package broker

// OSBClientResponse represents an OSB API response of a client service instance.
type OSBClientResponse struct {
	ServiceID  string                      `json:"service_id"`
	PlanID     string                      `json:"plan_id"`
	Parameters OSBClientResponseParameters `json:"parameters"`
}
type OSBClientResponseParameters struct {
	ClientID                  string   `json:"client_id"`
	ClientSecret              string   `json:"client_secret"`
	Protocol                  string   `json:"protocol"`
	PublicClient              bool     `json:"public_client,omitempty"`
	ClientAuthenticatorType   string   `json:"clientAuthenticatorType,omitempty"`
	RedirectURIs              []string `json:"redirect_uris"`
	WebOrigins                []string `json:"web_origins,omitempty"`
	StandardFlowEnabled       bool     `json:"standard_flow_enabled,omitempty"`
	ImplicitFlowEnabled       bool     `json:"implicit_flow_enabled,omitempty"`
	DirectAccessGrantsEnabled bool     `json:"direct_access_grants_enabled,omitempty"`
	ServiceAccountsEnabled    bool     `json:"service_accounts_enabled,omitempty"`
}

// OSBBindingResponse represents an OSB API response of a service instance binding.
type OSBBindingResponse struct {
	// TODO: ...
}
