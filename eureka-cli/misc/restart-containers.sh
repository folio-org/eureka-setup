#!/bin/bash

# Script to restart all Docker containers with "eureka-" prefix
echo "Finding containers with 'eureka-' prefix..."

# Get all container IDs that start with "eureka-"
CONTAINERS=$(docker ps -a --filter "name=^eureka-" --format "{{.ID}}")

if [ -z "$CONTAINERS" ]; then
  echo "No containers found with 'eureka-' prefix."
  exit 0
fi

echo "Found the following containers:"
docker ps -a --filter "name=^eureka-" --format "table {{.ID}}\t{{.Names}}\t{{.Status}}"

echo ""
echo "Restarting containers (2 at a time in parallel)..."

# Convert container list to array
CONTAINER_ARRAY=($CONTAINERS)
TOTAL=${#CONTAINER_ARRAY[@]}
INDEX=0

# Restart containers 2 at a time
while [ $INDEX -lt $TOTAL ]; do
  PIDS=()
  
  # Start up to 2 container restarts in background
  for i in 0 1; do
    if [ $INDEX -lt $TOTAL ]; then
      CONTAINER=${CONTAINER_ARRAY[$INDEX]}
      CONTAINER_NAME=$(docker inspect --format='{{.Name}}' $CONTAINER | sed 's/\///')
      echo "Restarting $CONTAINER_NAME ($CONTAINER)..."
      docker restart $CONTAINER &
      PIDS+=($!)
      INDEX=$((INDEX + 1))
    fi
  done
  
  # Wait for all restarts in this batch to complete
  for PID in "${PIDS[@]}"; do
    wait $PID
  done
done

echo ""
echo "All eureka- containers have been restarted successfully"
echo "Waiting 3 minutes for containers to stabilize..."
sleep 180
echo "Wait complete"
