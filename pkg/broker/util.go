package broker

import (
	"strings"

	"github.com/keycloak-broker/pkg/keycloak"
)

func keycloakClientToOSB(oidcClient *keycloak.OIDCClientResponse) OSBClientResponse {
	serviceData := strings.SplitN(oidcClient.Description, ";", 2)
	return OSBClientResponse{
		ServiceID: serviceData[0],
		PlanID:    serviceData[1],
		Parameters: OSBClientResponseParameters{
			ClientID:                  oidcClient.ClientID,
			ClientSecret:              oidcClient.Secret,
			Protocol:                  oidcClient.Protocol,
			PublicClient:              oidcClient.PublicClient,
			ClientAuthenticatorType:   oidcClient.ClientAuthenticatorType,
			RedirectURIs:              oidcClient.RedirectURIs,
			WebOrigins:                oidcClient.WebOrigins,
			StandardFlowEnabled:       oidcClient.StandardFlowEnabled,
			ImplicitFlowEnabled:       oidcClient.ImplicitFlowEnabled,
			DirectAccessGrantsEnabled: oidcClient.DirectAccessGrantsEnabled,
			ServiceAccountsEnabled:    oidcClient.ServiceAccountsEnabled,
		},
	}
}
