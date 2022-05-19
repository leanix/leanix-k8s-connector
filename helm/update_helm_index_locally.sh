#!/bin/bash

SPECIFIC_VERSION=$1
LATEST_BASE_VERSION=$2
set -e

# Creates or updates a new Helm chart package
# Specific semantic version
sed -i -E 's/(tag:).*/tag: '"$SPECIFIC_VERSION"'/g' leanix-k8s-connector/values.yaml
sed -i -E 's/(version:).*/version: '"$SPECIFIC_VERSION"'/g' leanix-k8s-connector/Chart.yaml
sed -i -E 's/(appVersion:).*/appVersion: '"\"$SPECIFIC_VERSION\""'/g' leanix-k8s-connector/Chart.yaml
helm package leanix-k8s-connector

# Latest version/chart
sed '-i.bak' -E 's/(tag:).*/tag: '"$LATEST_BASE_VERSION"'.latest/g' leanix-k8s-connector/values.yaml
helm package --version="$LATEST_BASE_VERSION".0.0-latest --app-version="$LATEST_BASE_VERSION".0.0-latest leanix-k8s-connector
rm leanix-k8s-connector/values.yaml
mv leanix-k8s-connector/values.yaml.bak leanix-k8s-connector/values.yaml

# Creates or updates the Helm chart repository index
helm repo index .