# LeanIX Kubernetes Connector Changelog

## Release 2021-09-20 - 5.0.0
* Breaking changes! Follow the below migration docs to upgrade from 4.0.0 to 5.0.0

### Migration docs
* New flag `enableCustomStorage` is introduced. This flags allows to disable the option to upload LDIF to custom storage backend. Disabling the flag will not affect the functionality of the connector.
* The default value is `false`. The new flag needs to be added with `true` value for same behaviour even after k8s connector upgrade to 5.0.0.
  `--set args.enableCustomStorage=true`
* `args.storageBackend` default is changed to `none`. Rename according to your configuration. Previously default value was `file`.
    e.g `--set args.enableCustomStorage=file`

## Release 2021-08-30 - 4.0.0

### Release Notes

* Breaking changes! Follow the below migration docs to upgrade from 3.0.0 or below to 4.0.0

#### Migration docs
* 3.0.0 or below to 4.0.0
  * Converted to a self-start connector of Integration Hub. Data source must be configured in the workspace before setting up the connector.
  * New **mandatory** flag is introduced to work with Integration hub data source - `integrationApi.datasourceName`
  * LDIF is still uploaded to choosen backend including the Integration Hub to trigger Integration API automatically. Hence `integrationApi.enabled` flag is removed
  * All the flags which are required when `integrationapi.enabled` is true should be passed such as `integrationApi.fqdn`, `integrationApi.secretName`
  * Integration API connector is automatically provisioned to the workspace. Dependency on cloud-beta is removed by introducing custom fields - `resolveStrategy`, `resolveLabel`
  * Integration API connector type is changed to `leanix-mi-connector` and connector id to `leanix-k8s-connector` hence the connector version is changed to `1.0.0`. The default value is also changed to `1.0.0` from `1.1.1`
  * `schedule.integrationApi` flag is removed and there is a single `schedule.standard`
  * Lowest possible value for `schedule.standard` is every hour


## Release 2021-08-04 - 3.0.0

### Release Notes

* Changes
  * Updated Integration API connector version and processor type default values to:
    * args.connectorVersion =  1.0.0 -> 1.1.1
    * args.processingMode = partial -> full

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
