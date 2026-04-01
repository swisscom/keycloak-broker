#!/bin/bash
set -euo pipefail

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

# client infos
echo "getting client [$CLIENT_ID] for [$CLIENT_ID] in realm [$REALM] ..."
export CLIENT_DATA=$(curl -s "$KEYCLOAK_URL/admin/realms/$REALM/clients?clientId=$CLIENT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
export CLIENT_UUID=$(echo $CLIENT_DATA | jq -r '.[0].id')
echo $CLIENT_DATA | jq .

echo "getting client secret for [$CLIENT_ID] in realm [$REALM] ..."
export CLIENT_SECRET=$(curl -s "$KEYCLOAK_URL/admin/realms/$REALM/clients/$CLIENT_UUID/client-secret" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.value')
echo $CLIENT_SECRET
