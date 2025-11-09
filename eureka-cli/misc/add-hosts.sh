#!/usr/bin/env bash

# Script to add Eureka hostnames to /etc/hosts
# This script is idempotent - it can be run multiple times safely
# Usage: sudo ./add-hosts.sh

set -e

# Define hostnames to add
HOSTNAMES=(
  "postgres.eureka"
  "kafka.eureka"
  "vault.eureka"
  "keycloak.eureka"
  "kong.eureka"
)

HOSTS_FILE="/etc/hosts"
IP_ADDRESS="127.0.0.1"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "Eureka Hosts Configuration Script"
echo "=================================="
echo ""

# Check if running with sufficient privileges
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Error: This script must be run with sudo or as root${NC}"
  echo "Usage: sudo $0"
  exit 1
fi

# Verify hosts file exists and is writable
if [ ! -f "$HOSTS_FILE" ]; then
  echo -e "${RED}Error: $HOSTS_FILE not found${NC}"
  exit 1
fi

if [ ! -w "$HOSTS_FILE" ]; then
  echo -e "${RED}Error: $HOSTS_FILE is not writable${NC}"
  exit 1
fi

# Create backup of hosts file
BACKUP_FILE="${HOSTS_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
cp "$HOSTS_FILE" "$BACKUP_FILE"
echo -e "${GREEN}Created backup: $BACKUP_FILE${NC}"
echo ""

# Track changes
ADDED_COUNT=0
SKIPPED_COUNT=0

# Process each hostname
for hostname in "${HOSTNAMES[@]}"; do
  # Check if hostname already exists in hosts file
  if grep -q "[[:space:]]${hostname}[[:space:]]*$" "$HOSTS_FILE" || \
     grep -q "[[:space:]]${hostname}$" "$HOSTS_FILE"; then
    echo -e "${YELLOW}⊖ Skipped${NC}: $hostname (already exists)"
    SKIPPED_COUNT=$((SKIPPED_COUNT + 1))
  else
    # Add entry to hosts file
    echo "$IP_ADDRESS $hostname" >> "$HOSTS_FILE"
    echo -e "${GREEN}✓ Added${NC}: $IP_ADDRESS $hostname"
    ADDED_COUNT=$((ADDED_COUNT + 1))
  fi
done

echo ""
echo "=================================="
echo "Summary:"
echo "  Added: $ADDED_COUNT"
echo "  Skipped: $SKIPPED_COUNT"
echo "  Backup: $BACKUP_FILE"
echo ""

if [ $ADDED_COUNT -gt 0 ]; then
  echo -e "${GREEN}Hosts file updated successfully!${NC}"
else
  echo -e "${YELLOW}No changes needed - all hostnames already configured${NC}"
fi

# Verify entries
echo ""
echo "Current Eureka entries in $HOSTS_FILE:"
echo "-----------------------------------"
for hostname in "${HOSTNAMES[@]}"; do
  grep "[[:space:]]${hostname}" "$HOSTS_FILE" || echo "(not found: $hostname)"
done
