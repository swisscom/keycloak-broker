#!/bin/bash

KEYCLOAK_URL="http://localhost:8080"
REALM="dev-realm"
ADMIN_USER="admin"
ADMIN_PASS="super-secret-admin-password"

docker kill keycloak-dev || true
docker rm keycloak-dev || true
docker run -d \
  --name keycloak-dev \
  -p 8080:8080 \
  -e KC_BOOTSTRAP_ADMIN_USERNAME=$ADMIN_USER \
  -e KC_BOOTSTRAP_ADMIN_PASSWORD=$ADMIN_PASS \
  quay.io/keycloak/keycloak:latest \
  start-dev

echo "Waiting for Keycloak to start..."
until curl -sf "$KEYCLOAK_URL/realms/master" > /dev/null 2>&1; do
  sleep 2
done

# Get the admin token
export ADMIN_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" \
  -d "username=$ADMIN_USER" \
  -d "password=$ADMIN_PASS" | jq -r '.access_token')

# Create the realm
curl -s -X POST "$KEYCLOAK_URL/admin/realms" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"realm": "'"$REALM"'", "enabled": true}'

echo "Keycloak dev server started at $KEYCLOAK_URL"
echo "Admin console: $KEYCLOAK_URL/admin (admin/super-secret-admin-password)"
echo "with realm [$REALM] already created ..."
