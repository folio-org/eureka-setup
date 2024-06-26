#!/bin/bash
# Wrapper script as docker entrypoint to run configure-realms.sh in parallel to actual kc.sh (the official entrypoint).

/opt/keycloak/bin/ebsco/configure-realms.sh &

/opt/keycloak/bin/kc.sh "$@"
