# LeanIX Kubernetes Connector Changelog

## Release 2021-06-21 - 2.0.0

### Release Notes

* Changes
  * Snyk - Update to alpine:3.11.6 to alpine:3.13.5 in Dockerfile
  * Snyk - Removed obsolete go dependencies

## Release 2020-10-22 - 2.0.0-beta7

### Release Notes

* Changes
  * Add `processingMode` field to generated LDIF file as this is required when processing mode in the Integration API processor configuration is set to `full`.

## Release 2020-10-21 - 2.0.0-beta6

### Release Notes

* Changes
  * Improved error logging for LeanIX Integration API integration

## Release 2020-10-15 - 2.0.0-beta5

### Release Notes

* New Features
  * Configure Kubernetes CronJob schedule
  * Upload LDIF to LeanIX Integration API

* Changes
  * The `connectorVersion` field does not contain the build version anymore. It is now configurable by the user to match the LeanIX Integration API processor configuration version.
  * The build version is moved to the section `customFields` and is mapped to the field `buildVersion` in the generated LDIF file.
  * Split LDIF and log upload into independent tasks
  * Use full container image path `docker.io/leanix/leanix-k8s-connector`

## Release 2020-06-15 - 2.0.0-beta4

### Release Notes

* Changes
  * Add `securityContext` section to the Helm chart.

    ```YAML
    securityContext:
      readOnlyRootFilesystem: true
      runAsNonRoot: true
      runAsUser: 65534
      runAsGroup: 65534
      allowPrivilegeEscalation: false
    ```

## Release 2020-04-28 - 2.0.0-beta3

### Release Notes

* New Features
  * Automatic container creation in Azure Storage, when using azureblob as storage backend

* Changes
  * Switch to Azure Storage Blob Go SDK v0.8.0
  * Switch from append blob to block blob

> **_NOTE:_** Delete all existing append blobs in the container in Azure Storage, you specified for the LDIF and log file upload. Otherwise the connector run throws an error, as append blobs cannot be overwritten with block blobs.

## Release 2020-02-07 - 2.0.0-beta2

### Release Notes

* New Features
  * Enable information extraction for the following Kubernetes resources:
    * replicasets
    * replicationcontrollers

* Changes
  * The `connectorId` field gets pinned to `Kubernetes` in the generated LDIF file.
  * The customer provided value for the `connectorId` field is moved to the section `customFields` and is mapped to the field `connectorInstance` in the generated LDIF file.

## Release 2020-01-14 - 2.0.0-beta1

### Release Notes

* New Features
  * Enable information extraction for the following Kubernetes resources:
    * serviceaccounts
    * services
    * nodes
    * pods
    * namespaces
    * configmaps
    * persistentvolumes
    * persistentvolumeclaims
    * deployments
    * statefulsets
    * daemonsets
    * customresourcedefinitions
    * clusterrolebindings
    * rolebindings
    * clusterroles
    * roles
    * ingresses
    * networkpolicies
    * horizontalpodautoscalers
    * podsecuritypolicies
    * storageclasses
    * cronjobs
    * jobs

* Removed Features
  * Custom Kubernetes resource extraction and information aggregation for:
    * deployments
    * statefulsets

* Changes
  * Permission change for the `leanix-k8s-connector` service account

## Release 2019-09-26 - 1.1.0

### Release Notes

* New Features
  * Enable wildcards for namespace filtering
