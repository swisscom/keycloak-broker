#!/bin/bash

KEYCLOAK_URL="http://localhost:8080"
REALM="dev-realm"
ADMIN_USER="admin"
ADMIN_PASS="super-secret-admin-password"
CLIENT_ID="9b788f00-6f12-402b-a2c9-b77a3edecb53" # to mimic a service-instance GUID
REDIRECT_URI="https://myapp.example.com/callback"

# Get the admin token
export ADMIN_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" \
  -d "username=$ADMIN_USER" \
  -d "password=$ADMIN_PASS" | jq -r '.access_token')

# Create the OIDC client
echo "creating client [$CLIENT_ID] in realm [$REALM] ..."
curl -X POST "$KEYCLOAK_URL/admin/realms/$REALM/clients" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "'"$CLIENT_ID"'",
    "name": "'"$CLIENT_ID"'",
    "description": "CorpIDv3 client",
    "enabled": true,
    "protocol": "openid-connect",
    "publicClient": false,
    "redirectUris": ["'"$REDIRECT_URI"'"],
    "standardFlowEnabled": true,
    "directAccessGrantsEnabled": false,
    "serviceAccountsEnabled": false
  }'

# client secret
echo "getting client secret for [$CLIENT_ID] in realm [$REALM] ..."
export CLIENT_UUID=$(curl -s "$KEYCLOAK_URL/admin/realms/$REALM/clients?clientId=$CLIENT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')
export RESPONSE=$(curl -s "$KEYCLOAK_URL/admin/realms/$REALM/clients/$CLIENT_UUID/client-secret" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
#echo $RESPONSE
export CLIENT_SECRET=$(echo $RESPONSE | jq -r '.value')
echo $CLIENT_SECRET
