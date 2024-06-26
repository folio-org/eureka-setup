#!/usr/bin/env bash

config_path="/opt/kong/config"
export KONG_PLUGINS="$KONG_PLUGINS,auth-headers-manager"

# Bootstrap Kong
kong migrations bootstrap
kong migrations up
kong migrations finish

# Start Kong
kong start

# Wait for Kong to start up
until curl -s http://localhost:8001 >/dev/null 2>&1; do
  echo "Waiting for Kong to start..."
  sleep 1
done

echo "Kong initialization..."

# Set mgr urls for UI
sed -i s#\{env}#"${ENV}"#g $config_path/ui.yaml

# Get the names of all files in the config directory
config_files=$(ls $config_path)

# Create the deck command with each file as a separate -s option
deck_cmd="deck sync --select-tag=automation"
for file in $config_files; do
  deck_cmd="$deck_cmd -s $config_path/$file"
done

# Run the deck command
echo "$deck_cmd"
$deck_cmd

echo "Kong initialization finished successfully!"

# Stop Kong
kong stop

source /docker-entrypoint.sh "$@"
