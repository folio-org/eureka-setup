#!/bin/bash

set -e

echo "Started Git cloning modules"

WORK_DIR="${1:-$HOME/Folio/eureka-setup/local}"
MODULE_DIR="${2:-cloned-modules}"
MODULE_JSON="${3:-cloneable-modules.json}"

mkdir -p "$WORK_DIR/$MODULE_DIR"

while IFS== read key value; do
  MODULE_NAME=$(echo $value | jq -r .name)
  MODULE_URL=$(echo $value | jq -r .url)

  echo "Attempting to clone $MODULE_NAME module"
  
  git clone "$MODULE_URL" --branch master "$WORK_DIR/$MODULE_DIR/$MODULE_NAME" || true

  cd "$WORK_DIR/$MODULE_DIR/$MODULE_NAME" || exit

  git pull origin master

  mvn clean install -DskipTests 

  cd "$WORK_DIR/$MODULE_DIR" || exit
done < <(jq -r 'to_entries|map("\(.key)=\(.value|tostring)")|.[]' $WORK_DIR/$MODULE_JSON)

echo "Stopped Git cloning modules"