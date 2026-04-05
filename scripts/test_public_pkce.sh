#!/bin/bash
set -euo pipefail

BROKER_URL="http://disco:dingo@localhost:9999"
KEYCLOAK_URL="http://localhost:8080"
REALM="dev-realm"
INSTANCE_ID="a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
BINDING_ID="b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e"
SERVICE_ID="fff5b36a-da19-4dc2-bd28-3dd331146290"
PLAN_ID="ece8f4ad-1ba0-481b-94b0-8fb4f066aa38" # public-client
USERNAME="einstein@example.com"
PASSWORD="relative"

cleanup() {
  echo ""
  echo "=== Cleanup: deprovisioning instance ==="
  curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID" -X DELETE || true
}
trap cleanup EXIT

echo "=== Provisioning public client with PKCE ==="
curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID" \
  -X PUT -H "Content-Type: application/json" \
  -d '{
    "service_id": "'"$SERVICE_ID"'",
    "plan_id": "'"$PLAN_ID"'",
    "parameters": {
      "redirectURIs": ["https://myapp.example.com/callback"],
      "pkceEnabled": true,
      "directAccessGrantsEnabled": true,
      "refreshTokenLifetime": 699,
      "accessTokenLifetime": 333
    }
  }' | jq .

echo ""
echo "=== Fetching instance ==="
INSTANCE=$(curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID")
echo "$INSTANCE" | jq .

# verify PKCE is enabled
PKCE=$(echo "$INSTANCE" | jq -r '.parameters.pkceEnabled')
if [ "$PKCE" != "true" ]; then
  echo "FAIL: pkceEnabled expected true, got $PKCE"
  exit 1
fi
echo "PASS: pkceEnabled is true"

# verify it's a public client
PUBLIC=$(echo "$INSTANCE" | jq -r '.parameters.publicClient')
if [ "$PUBLIC" != "true" ]; then
  echo "FAIL: publicClient expected true, got $PUBLIC"
  exit 1
fi
echo "PASS: publicClient is true"

# verify refresh token lifetime
RTL=$(echo "$INSTANCE" | jq -r '.parameters.refreshTokenLifetime')
if [ "$RTL" != "699" ]; then
  echo "FAIL: refreshTokenLifetime expected 699, got $RTL"
  exit 1
fi
echo "PASS: refreshTokenLifetime is 699"

# verify access token lifetime
ATL=$(echo "$INSTANCE" | jq -r '.parameters.accessTokenLifetime')
if [ "$ATL" != "333" ]; then
  echo "FAIL: accessTokenLifetime expected 333, got $ATL"
  exit 1
fi
echo "PASS: accessTokenLifetime is 333"

echo ""
echo "=== Creating binding ==="
BINDING=$(curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID/service_bindings/$BINDING_ID" \
  -X PUT -H "Content-Type: application/json")
echo "$BINDING" | jq .

CLIENT_ID=$(echo "$BINDING" | jq -r '.credentials.clientId')
CLIENT_SECRET=$(echo "$BINDING" | jq -r '.credentials.clientSecret')
echo "Client ID: $CLIENT_ID"

if [ -n "$CLIENT_SECRET" ]; then
  echo "FAIL: clientSecret expected empty for public client, got $CLIENT_SECRET"
  exit 1
fi
echo "PASS: clientSecret is empty"

echo ""
echo "=== Testing OIDC login (direct access grant, public client, no secret) ==="
TOKEN_RESPONSE=$(curl -sf -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
  -d "grant_type=password" \
  -d "client_id=$CLIENT_ID" \
  -d "username=$USERNAME" \
  -d "password=$PASSWORD" \
  -d "scope=openid email phone")
ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
echo "Access token obtained"

echo ""
echo "=== Fetching userinfo ==="
curl -sf "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/userinfo" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .

echo ""
echo "=== Deleting binding ==="
curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID/service_bindings/$BINDING_ID" -X DELETE | jq .

echo ""
echo "ALL TESTS PASSED"
