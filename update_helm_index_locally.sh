#!/bin/bash

LATEST_BASE_VERSION=6
set -e

cd ./helm

# Creates or updates a new Helm chart package
# Specific semantic version
helm package ./leanix-k8s-connector

# Latest version/chart
sed -i '.bak' -E 's/(tag:).*/tag: '"$LATEST_BASE_VERSION"'.latest/g' ./leanix-k8s-connector/values.yaml
helm package --version="$LATEST_BASE_VERSION".0.0-latest --app-version="$LATEST_BASE_VERSION".0.0-latest ./leanix-k8s-connector
rm leanix-k8s-connector/values.yaml
mv leanix-k8s-connector/values.yaml.bak leanix-k8s-connector/values.yaml

# Creates or updates the Helm chart repository index
helm repo index .