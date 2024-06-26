#!/bin/bash

script="configure-realms.sh"

keycloakUrl="http://localhost:8080"
clientId="superAdmin"
clientSecret="superAdmin"

maxAttempts=50
attemptCounter=0

function loginAsAdmin() {
  /opt/keycloak/bin/kcadm.sh config credentials \
    --server "$keycloakUrl" \
    --realm master \
    --user admin \
    --password admin \
    &> /dev/null
}

while [ $attemptCounter -le $maxAttempts ]; do
  echo "$(date +%F' '%T,%3N) INFO  [$script] Trying to add client: '$clientId' to master realm [attempt: $attemptCounter]"

  if loginAsAdmin; then
    /opt/keycloak/bin/ebsco/setup-admin-client.sh "$clientId" "$clientSecret"
    break
  fi

  echo "$(date +%F' '%T,%3N) INFO  [$script] Keycloak is not ready yet, waiting for 10 seconds [attempt: $attemptCounter]"

  attemptCounter=$((attemptCounter + 1))
  
  sleep 10
done

if [ $attemptCounter -ge $maxAttempts ]; then
  echo "$(date +%F' '%T,%3N) WARN  [$script] Failed to add client: $clientId to master realm, the amount of attempt is greater than $maxAttempts"
  exit 1
fi
