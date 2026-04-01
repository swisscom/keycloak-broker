#!/bin/bash
set -euo pipefail

KEYCLOAK_URL="http://localhost:8080"
REALM="dev-realm"
BROKER_URL="http://disco:dingo@localhost:9999"
INSTANCE_ID="fe5556b9-8478-409b-ab2b-3c95ba06c5fc"
USERNAME="einstein@example.com"
PASSWORD="relative"

# Fetch client credentials from the broker
echo "Fetching client credentials from broker ..."
BINDING=$(curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID/service_bindings/db59931a-70a6-43c1-8885-b0c6b1c194d4")
CLIENT_ID=$(echo "$BINDING" | jq -r '.credentials.clientId')
CLIENT_SECRET=$(echo "$BINDING" | jq -r '.credentials.clientSecret')
echo "Client ID:     $CLIENT_ID"
echo "Client Secret: $CLIENT_SECRET"
echo ""

# Login with test user via Resource Owner Password Credentials grant
echo "Logging in as $USERNAME ..."
TOKEN_RESPONSE=$(curl -sf -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
  -d "grant_type=password" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "username=$USERNAME" \
  -d "password=$PASSWORD" \
  -d "scope=openid")
ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
echo "Access token obtained: [$ACCESS_TOKEN]"
echo ""

# Call userinfo endpoint
echo "Fetching userinfo ..."
curl -sf "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/userinfo" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
