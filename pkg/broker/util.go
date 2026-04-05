package broker

import (
	"strconv"

	"github.com/keycloak-broker/pkg/keycloak"
)

func keycloakClientToOSB(oidcClient *keycloak.OIDCClientResponse) OSBClientResponse {
	pkceEnabled := oidcClient.Attributes["pkce.code.challenge.method"] == "S256"
	refreshTokenLifespan, _ := strconv.Atoi(oidcClient.Attributes["client.session.max.lifespan"])

	return OSBClientResponse{
		ServiceId: oidcClient.Attributes["service_id"],
		PlanId:    oidcClient.Attributes["plan_id"],
		Parameters: OSBClientResponseParameters{
			ClientId:                  oidcClient.ClientId,
			ClientSecret:              oidcClient.Secret,
			Description:               oidcClient.Description,
			Protocol:                  oidcClient.Protocol,
			PublicClient:              oidcClient.PublicClient,
			ClientAuthenticatorType:   oidcClient.ClientAuthenticatorType,
			RedirectURIs:              oidcClient.RedirectURIs,
			WebOrigins:                oidcClient.WebOrigins,
			ConsentRequired:           oidcClient.ConsentRequired,
			StandardFlowEnabled:       oidcClient.StandardFlowEnabled,
			ImplicitFlowEnabled:       oidcClient.ImplicitFlowEnabled,
			DirectAccessGrantsEnabled: oidcClient.DirectAccessGrantsEnabled,
			ServiceAccountsEnabled:    oidcClient.ServiceAccountsEnabled,
			PKCEEnabled:               pkceEnabled,
			RefreshTokenLifespan:      refreshTokenLifespan,
			Issuer:                    oidcClient.Issuer,
			DiscoveryEndpoint:         oidcClient.DiscoveryEndpoint,
			AuthorizationEndpoint:     oidcClient.AuthorizationEndpoint,
			TokenEndpoint:             oidcClient.TokenEndpoint,
			IntrospectionEndpoint:     oidcClient.IntrospectionEndpoint,
			UserInfoEndpoint:          oidcClient.UserInfoEndpoint,
			EndSessionEndpoint:        oidcClient.EndSessionEndpoint,
			JWKSURI:                   oidcClient.JWKSURI,
		},
	}
}

func keycloakClientToOSBBinding(oidcClient *keycloak.OIDCClientResponse) OSBBindingResponse {
	return OSBBindingResponse{
		Metadata: OSBBindingResponseMetadata{
			ServiceId: oidcClient.Attributes["service_id"],
			PlanId:    oidcClient.Attributes["plan_id"],
		},
		Credentials: OSBBindingResponseCredentials{
			ClientId:              oidcClient.ClientId,
			ClientSecret:          oidcClient.Secret,
			RedirectURIs:          oidcClient.RedirectURIs,
			Issuer:                oidcClient.Issuer,
			DiscoveryEndpoint:     oidcClient.DiscoveryEndpoint,
			AuthorizationEndpoint: oidcClient.AuthorizationEndpoint,
			TokenEndpoint:         oidcClient.TokenEndpoint,
			IntrospectionEndpoint: oidcClient.IntrospectionEndpoint,
			UserInfoEndpoint:      oidcClient.UserInfoEndpoint,
			EndSessionEndpoint:    oidcClient.EndSessionEndpoint,
			JWKSURI:               oidcClient.JWKSURI,
		},
	}
}
