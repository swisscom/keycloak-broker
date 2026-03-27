package broker

import (
	"github.com/keycloak-broker/pkg/keycloak"
)

func keycloakClientToOSB(oidcClient *keycloak.OIDCClientResponse) OSBClientResponse {
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
		},
	}
}
