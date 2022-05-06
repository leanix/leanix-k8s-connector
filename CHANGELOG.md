# LeanIX Kubernetes Connector Changelog

## Release 2022-05-05 - 6.6.0
* Added relation between software artifacts and deployments

## Release 2022-04-29 - 6.5.2
* Added software Artifact mapping to beta version

## Release 2022-04-26 - 6.5.1
* Update the base image for vulnerability fix. More details [here](https://security.snyk.io/vuln/SNYK-ALPINE313-BUSYBOX-2440609)

## Release 2022-04-08 - 6.5.0
* Upgraded to new iHub URL for self start runs.

## Release 2022-03-16 - 6.4.1
* Added error scopes to prevent crashes and also extended the logging for integration hub connection

## Release 2022-03-09 - 6.4.0
* Added beta version access to in development k8s-connector version. Version usage is configured in the k8s integration page in the workspace.

## Release 2022-03-07 - 6.3.3
* Increased logging and added more precise log messages

## Release 2022-01-20 - 6.3.2
* Error messages can be captured and accessed via connector logs in iHub, if the connector configuration is not correct.

## Release 2022-01-10 - 6.3.1
* All k8s workloads are now prefixed with the Helm release name instead of being suffixed to avoid length issues with kubernetes.

## Release 2022-01-03 - 6.3.0
* Connector Logs can be accessed in iHub (w.r.t specific dataSource name) `Show Log` button for troubleshooting and monitoring purpose.

## Release 2022-01-03 - 6.2.0
* To ensure support for multiple helm release on same cluster, all the existing K8s workloads are suffixed with the Helm release name.
* Support multiple workspace ID for same cluster via different helm releases

## Release 2021-12-08 - 6.1.0
* Breaking changes! Follow the below migration docs to upgrade from 6.0.3 to 6.1.0

### Migration docs
* `clustername` `connectorID` `connectorVersion` `processingMode` can not be set by helm parameters anymore.
* The iHUB connector configuration for the k8s integration is extended to capture the 'clusterName'.

## Release 2021-12-06 - 6.0.3
* Dependencies have been updated

## Release 2021-11-19 - 6.0.2
* Empty `connectorID` will now use the `clustername` instead of a random uuid.

## Release 2021-11-18 - 6.0.1
* Improved error logging for LeanIX Integration Hub API calls

## Release 2021-10-12 - 6.0.0
* Breaking changes! Adapted changes for VSM data model. No changes in the behaviour of the connector.
* No backward compatibility with MI data model

## Release 2021-09-20 - 5.0.0
* Breaking changes! Follow the below migration docs to upgrade from 4.0.0 to 5.0.0

### Migration docs
* New flag `enableCustomStorage` is introduced. This flags allows to disable the option to upload LDIF to custom storage backend. Disabling the flag will not affect the functionality of the connector.
* The default value is `false`. The new flag needs to be added with `true` value for same behaviour even after k8s connector upgrade to 5.0.0.
  `--set args.enableCustomStorage=true`
* `args.storageBackend` default is changed to `none`. Rename according to your configuration. Previously default value was `file`.
    e.g `--set args.storageBackend=file`

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
