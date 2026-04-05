#!/bin/bash
set -euo pipefail

KEYCLOAK_URL="http://localhost:8080"
REALM="dev-realm"
BROKER_URL="http://disco:dingo@localhost:9999"
INSTANCE_ID="fe5556b9-8478-409b-ab2b-3c95ba06c5fc"
BINDING_ID="db59931a-70a6-43c1-8885-b0c6b1c194d4"
SERVICE_ID="fff5b36a-da19-4dc2-bd28-3dd331146290"
PLAN_ID="40627d0f-dedd-4d68-8111-2ebae510ba1b"
USERNAME="einstein@example.com"
PASSWORD="relative"

cleanup() {
  echo ""
  echo "=== Cleanup ==="
  curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID/service_bindings/$BINDING_ID" -X DELETE || true
  curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID" -X DELETE || true
}
trap cleanup EXIT

# ==================== Provision ====================
echo "=== Provisioning service instance ==="
PROVISION=$(curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID" \
  -X PUT -H "Content-Type: application/json" \
  -d '{
    "service_id": "'"$SERVICE_ID"'",
    "plan_id": "'"$PLAN_ID"'",
    "parameters": {
      "redirectURIs": ["https://myapp.example.com/callback"],
      "directAccessGrantsEnabled": true
    }
  }')
echo "$PROVISION" | jq .

PROV_CLIENT_ID=$(echo "$PROVISION" | jq -r '.parameters.clientId')
if [ "$PROV_CLIENT_ID" != "$INSTANCE_ID" ]; then
  echo "FAIL: provision clientId = $PROV_CLIENT_ID, want $INSTANCE_ID"
  exit 1
fi
echo "PASS: provision clientId is correct"

PROV_PKCE=$(echo "$PROVISION" | jq -r '.parameters.pkceEnabled')
if [ "$PROV_PKCE" != "true" ]; then
  echo "FAIL: provision pkceEnabled = $PROV_PKCE, want true (default)"
  exit 1
fi
echo "PASS: provision pkceEnabled defaults to true"

PROV_STD=$(echo "$PROVISION" | jq -r '.parameters.standardFlowEnabled')
if [ "$PROV_STD" != "true" ]; then
  echo "FAIL: provision standardFlowEnabled = $PROV_STD, want true"
  exit 1
fi
echo "PASS: provision standardFlowEnabled is true"

# ==================== Fetch instance ====================
echo ""
echo "=== Fetching service instance ==="
INSTANCE=$(curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID")
echo "$INSTANCE" | jq .

FETCH_CLIENT_ID=$(echo "$INSTANCE" | jq -r '.parameters.clientId')
if [ "$FETCH_CLIENT_ID" != "$INSTANCE_ID" ]; then
  echo "FAIL: fetch clientId = $FETCH_CLIENT_ID, want $INSTANCE_ID"
  exit 1
fi
echo "PASS: fetch clientId is correct"

FETCH_REDIRECT=$(echo "$INSTANCE" | jq -r '.parameters.redirectURIs[0]')
if [ "$FETCH_REDIRECT" != "https://myapp.example.com/callback" ]; then
  echo "FAIL: fetch redirectURIs[0] = $FETCH_REDIRECT, want https://myapp.example.com/callback"
  exit 1
fi
echo "PASS: fetch redirectURIs is correct"

# ==================== Update instance ====================
echo ""
echo "=== Updating service instance ==="
UPDATE=$(curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID" \
  -X PATCH -H "Content-Type: application/json" \
  -d '{
    "service_id": "'"$SERVICE_ID"'",
    "plan_id": "'"$PLAN_ID"'",
    "parameters": {
      "redirectURIs": ["https://myapp.example.com/callback", "https://other.example.com/callback"],
      "implicitFlowEnabled": true
    }
  }')
echo "$UPDATE" | jq .

UPD_REDIRECT_COUNT=$(echo "$UPDATE" | jq '.parameters.redirectURIs | length')
if [ "$UPD_REDIRECT_COUNT" != "2" ]; then
  echo "FAIL: update redirectURIs count = $UPD_REDIRECT_COUNT, want 2"
  exit 1
fi
echo "PASS: update redirectURIs has 2 entries"

UPD_IMPLICIT=$(echo "$UPDATE" | jq -r '.parameters.implicitFlowEnabled')
if [ "$UPD_IMPLICIT" != "true" ]; then
  echo "FAIL: update implicitFlowEnabled = $UPD_IMPLICIT, want true"
  exit 1
fi
echo "PASS: update implicitFlowEnabled is true"

UPD_DAG=$(echo "$UPDATE" | jq -r '.parameters.directAccessGrantsEnabled')
if [ "$UPD_DAG" != "true" ]; then
  echo "FAIL: update directAccessGrantsEnabled = $UPD_DAG, want true (preserved)"
  exit 1
fi
echo "PASS: update directAccessGrantsEnabled preserved as true"

# ==================== Create binding ====================
echo ""
echo "=== Creating binding ==="
BINDING=$(curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID/service_bindings/$BINDING_ID" \
  -X PUT -H "Content-Type: application/json")
echo "$BINDING" | jq .

CLIENT_ID=$(echo "$BINDING" | jq -r '.credentials.clientId')
CLIENT_SECRET=$(echo "$BINDING" | jq -r '.credentials.clientSecret')
TOKEN_ENDPOINT=$(echo "$BINDING" | jq -r '.credentials.tokenEndpoint')
ISSUER=$(echo "$BINDING" | jq -r '.credentials.issuer')

if [ -z "$CLIENT_ID" ] || [ "$CLIENT_ID" = "null" ]; then
  echo "FAIL: binding clientId is empty"
  exit 1
fi
echo "PASS: binding clientId is $CLIENT_ID"

if [ -z "$CLIENT_SECRET" ] || [ "$CLIENT_SECRET" = "null" ]; then
  echo "FAIL: binding clientSecret is empty"
  exit 1
fi
echo "PASS: binding clientSecret is present"

if [ -z "$TOKEN_ENDPOINT" ] || [ "$TOKEN_ENDPOINT" = "null" ]; then
  echo "FAIL: binding tokenEndpoint is empty"
  exit 1
fi
echo "PASS: binding tokenEndpoint is $TOKEN_ENDPOINT"

if [ -z "$ISSUER" ] || [ "$ISSUER" = "null" ]; then
  echo "FAIL: binding issuer is empty"
  exit 1
fi
echo "PASS: binding issuer is $ISSUER"

# ==================== Fetch binding ====================
echo ""
echo "=== Fetching binding ==="
FETCH_BINDING=$(curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID/service_bindings/$BINDING_ID")
echo "$FETCH_BINDING" | jq .

FB_CLIENT_ID=$(echo "$FETCH_BINDING" | jq -r '.credentials.clientId')
if [ "$FB_CLIENT_ID" != "$CLIENT_ID" ]; then
  echo "FAIL: fetch binding clientId = $FB_CLIENT_ID, want $CLIENT_ID"
  exit 1
fi
echo "PASS: fetch binding clientId matches"

# ==================== OIDC login ====================
echo ""
echo "=== Logging in as $USERNAME ==="
TOKEN_RESPONSE=$(curl -sf -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
  -d "grant_type=password" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "username=$USERNAME" \
  -d "password=$PASSWORD" \
  -d "scope=openid email phone")

ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
  echo "FAIL: access_token is empty"
  exit 1
fi
echo "PASS: access token obtained"

REFRESH_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.refresh_token')
if [ -z "$REFRESH_TOKEN" ] || [ "$REFRESH_TOKEN" = "null" ]; then
  echo "FAIL: refresh_token is empty"
  exit 1
fi
echo "PASS: refresh token obtained"

TOKEN_TYPE=$(echo "$TOKEN_RESPONSE" | jq -r '.token_type')
if [ "$TOKEN_TYPE" != "Bearer" ]; then
  echo "FAIL: token_type = $TOKEN_TYPE, want Bearer"
  exit 1
fi
echo "PASS: token_type is Bearer"

# ==================== Userinfo ====================
echo ""
echo "=== Fetching userinfo ==="
USERINFO=$(curl -sf "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/userinfo" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "$USERINFO" | jq .

EMAIL=$(echo "$USERINFO" | jq -r '.email')
if [ "$EMAIL" != "$USERNAME" ]; then
  echo "FAIL: email = $EMAIL, want $USERNAME"
  exit 1
fi
echo "PASS: email is $EMAIL"

SUB=$(echo "$USERINFO" | jq -r '.sub')
if [ -z "$SUB" ] || [ "$SUB" = "null" ]; then
  echo "FAIL: sub is empty"
  exit 1
fi
echo "PASS: sub is $SUB"

# ==================== Delete binding ====================
echo ""
echo "=== Deleting binding ==="
curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID/service_bindings/$BINDING_ID" -X DELETE | jq .
echo "PASS: binding deleted"

# ==================== Deprovision ====================
echo ""
echo "=== Deprovisioning service instance ==="
curl -sf "$BROKER_URL/v2/service_instances/$INSTANCE_ID" -X DELETE | jq .
echo "PASS: instance deprovisioned"

# verify instance is gone
echo ""
echo "=== Verifying instance is gone ==="
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BROKER_URL/v2/service_instances/$INSTANCE_ID")
if [ "$HTTP_CODE" != "404" ]; then
  echo "FAIL: expected 404 after deprovision, got $HTTP_CODE"
  exit 1
fi
echo "PASS: instance returns 404"

echo ""
echo "ALL TESTS PASSED"
