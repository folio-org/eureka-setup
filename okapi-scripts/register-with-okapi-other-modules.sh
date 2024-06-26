#!/bin/bash

set -e

# Keycloak
MOD_LOGIN_KEYCLOAK=1.0.0
MOD_USERS_KEYCLOAK=1.0.0
MOD_ROLES_KEYCLOAK=1.0.0
MOD_SCHEDULER_VERSION=1.0.0
# Core
MOD_AUTHTOKEN_VERSION=2.15.1
MOD_LOGIN_VERSION=7.11.0
MOD_USERS_VERSION=19.3.1
MOD_PERMISSIONS_VERSION=6.5.0
MOD_CONFIGURATION_VERSION=5.10.0
# Others
MOD_PASSWORD_VALIDATOR_VERSION=3.2.0
MOD_TAGS_VERSION=2.2.0
MOD_USERS_BL_VERSION=7.7.0
MOD_NOTES_VERSION=5.2.0
MOD_SETTINGS_VERSION=1.0.3
MOD_INVENTORY_STORAGE_VERSION=27.1.0
MOD_PUBSUB_VERSION=2.13.0
MOD_CIRCULATION_STORAGE_VERSION=17.2.0

# Keycloak 
# WARNING: Disabled, require additional module interfaces
# # mod-login-keycloak
# ./okapi-scripts/enable-module.sh \
#   --okapi "http://localhost:9130" \
#   --username "superuser" \
#   --password "superAdmin" \
#   --login \
#   --module "mod-login-keycloak" \
#   --version "$MOD_LOGIN_KEYCLOAK" \
#   --module-url "http://mod-login-keycloak.eureka:8081"
# # mod-users-keycloak
#  ./okapi-scripts/enable-module.sh \
#    --okapi "http://localhost:9130" \
#    --username "superuser" \
#    --password "superAdmin" \
#    --login \
#    --module "mod-users-keycloak" \
#    --version "$MOD_USERS_KEYCLOAK" \
#    --module-url "http://mod-users-keycloak.eureka:8081"
# # mod-roles-keycloak
# ./okapi-scripts/enable-module.sh \
#   --okapi "http://localhost:9130" \
#   --username "superuser" \
#   --password "superAdmin" \
#   --login \
#   --module "mod-roles-keycloak" \
#   --version "$MOD_ROLES_KEYCLOAK" \
#   --module-url "http://mod-roles-keycloak.eureka:8081"
# # mod-scheduler
# ./okapi-scripts/enable-module.sh \
#   --okapi "http://localhost:9130" \
#   --username "superuser" \
#   --password "superAdmin" \
#   --login \
#   --module "mod-scheduler" \
#   --version "$MOD_SCHEDULER_VERSION" \
#   --module-url "http://mod-scheduler.eureka:8081"

# Core
# mod-configuration
./okapi-scripts/enable-module.sh \
  --okapi "http://localhost:9130" \
  --username "superuser" \
  --password "superAdmin" \
  --login \
  --module "mod-configuration" \
  --version "$MOD_CONFIGURATION_VERSION" \
  --module-url "http://mod-configuration.eureka:8081"

# Others
# mod-password-validator
./okapi-scripts/enable-module.sh \
  --okapi "http://localhost:9130" \
  --username "superuser" \
  --password "superAdmin" \
  --login \
  --module "mod-password-validator" \
  --version "$MOD_PASSWORD_VALIDATOR_VERSION" \
  --module-url "http://mod-password-validator.eureka:8081"
# mod-tags
./okapi-scripts/enable-module.sh \
  --okapi "http://localhost:9130" \
  --username "superuser" \
  --password "superAdmin" \
  --login \
  --module "mod-tags" \
  --version "$MOD_TAGS_VERSION" \
  --module-url "http://mod-tags.eureka:8081"
# mod-users-bl
./okapi-scripts/enable-module.sh \
  --okapi "http://localhost:9130" \
  --username "superuser" \
  --password "superAdmin" \
  --login \
  --module "mod-users-bl" \
  --version "$MOD_USERS_BL_VERSION" \
  --module-url "http://mod-users-bl.eureka:8081"
# mod-notes
./okapi-scripts/enable-module.sh \
  --okapi "http://localhost:9130" \
  --username "superuser" \
  --password "superAdmin" \
  --login \
  --module "mod-notes" \
  --version "$MOD_NOTES_VERSION" \
  --module-url "http://mod-notes.eureka:8081"
# mod-settings
./okapi-scripts/enable-module.sh \
  --okapi "http://localhost:9130" \
  --username "superuser" \
  --password "superAdmin" \
  --login \
  --module "mod-settings" \
  --version "$MOD_SETTINGS_VERSION" \
  --module-url "http://mod-settings.eureka:8081"
# mod-inventory-storage
./okapi-scripts/enable-module.sh \
  --okapi "http://localhost:9130" \
  --username "superuser" \
  --password "superAdmin" \
  --login \
  --module "mod-inventory-storage" \
  --version "$MOD_INVENTORY_STORAGE_VERSION" \
  --module-url "http://mod-inventory-storage.eureka:8081"
# mod-pubsub
./okapi-scripts/enable-module.sh \
  --okapi "http://localhost:9130" \
  --username "superuser" \
  --password "superAdmin" \
  --login \
  --module "mod-pubsub" \
  --version "$MOD_PUBSUB_VERSION" \
  --module-url "http://mod-pubsub.eureka:8081"
# mod-circulation-storage
./okapi-scripts/enable-module.sh \
  --okapi "http://localhost:9130" \
  --username "superuser" \
  --password "superAdmin" \
  --login \
  --module "mod-circulation-storage" \
  --version "$MOD_CIRCULATION_STORAGE_VERSION" \
  --module-url "http://mod-circulation-storage.eureka:8081"
