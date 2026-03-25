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

# Delete the OIDC client
echo "getting internal client_uuid for [$CLIENT_ID] in realm [$REALM] ..."
export CLIENT_UUID=$(curl -s "$KEYCLOAK_URL/admin/realms/$REALM/clients?clientId=$CLIENT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')
echo "deleting client [$CLIENT_ID] with client_uuid [$CLIENT_UUID] in realm [$REALM] ..."
curl -X DELETE "$KEYCLOAK_URL/admin/realms/$REALM/clients/$CLIENT_UUID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
echo "client [$CLIENT_ID] deleted ..."

# test
curl -s "$KEYCLOAK_URL/admin/realms/$REALM/clients?clientId=$CLIENT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
