# LeanIX Kubernetes Connector

**Note**: This document is intended solely for developers. MI users should follow the [LeanIX MI docs](https://docs-mi.leanix.net/docs/kubernetes).

The LeanIX Kubernetes Connector collects information from Kubernetes.

## Table of contents

- [LeanIX Kubernetes Connector](#leanix-kubernetes-connector)
  - [Table of contents](#table-of-contents)
  - [Overview](#overview)
  - [Getting started](#getting-started)
    - [Integration Hub](#integration-hub)
    - [Architecture](#architecture)
    - [Installation - Helm chart](#installation---helm-chart)
      - [Pre-release versions](#pre-release-versions)
      - [Helm 2 requirements](#helm-2-requirements)
      - [Add LeanIX Kubernetes Connector Helm chart repository](#add-leanix-kubernetes-connector-helm-chart-repository)
      - [Starting Connector in k8s](#starting-connector-in-k8s)
      - [file storage backend](#file-storage-backend)
      - [azureblob storage backend](#azureblob-storage-backend)
      - [Optional - Advanced deployment settings](#optional---advanced-deployment-settings)
    - [Setting up development environment](#developer-environment-setup)
  - [Known issues](#known-issues)
  - [Version history](#version-history)

## Overview

The LeanIX Kubernetes Connector (Integration Hub self start) runs in the Kubernetes cluster as a container itself and collects information from the cluster like namespaces, deployments, pods, etc. The information are sanitized and brought into the LDIF (LeanIX Data Interchange Format) format that LeanIX understands. The output then is stored in the `kubernetes.ldif` file that gets imported into LeanIX. The progress is updated to Integration Hub datasource. The Integration API run is taken care by Integration Hub.

## Getting started

Depending on how you would like to run the LeanIX Kubernetes Connector the installation steps differ a bit and depends on the selected storage backend.

### Integration Hub

#### Connector template

- mi-k8s-connector

### Architecture

The LeanIX Kubernetes Connector gets deployed via a Helm chart into the Kubernetes cluster as a CronJob. All necessary requirements like the ServiceAccount, the ClusterRole and ClusterRoleBinding are deployed also by the Helm chart.

Only necessary permissions are given to the connector as the default ClusterRole `view` and additional permissions listed below that are not part of the ClusterRole `view`.

- [Kubernetes ClusterRole view](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#user-facing-roles)

| apiGroups                   | resources                                              | verbs            |
| --------------------------- | ------------------------------------------------------ | ---------------- |
| ""                          | nodes, persistentvolumes                               | get, list, watch |
| "apiextensions.k8s.io"      | customresourcedefinitions                              | get, list, watch |
| "policy"                    | podsecuritypolicies                                    | get, list, watch |
| "rbac.authorization.k8s.io" | roles, clusterroles, rolebindings, clusterrolebindings | get, list, watch |
| "storage.k8s.io"            | storageclasses                                         | get, list, watch |

The CronJob is configured to run every hour and spins up a new pod of the LeanIX Kubernetes Connector. As mentioned in the overview the connector creates the `kubernetes.ldif` file and logs into the `leanix-k8s-connector.log` file.

Currently, two storage backend types are natively supported by the connector.

- file
- azureblob

The `file` storage backend lets you use every storage that can be provided to Kubernetes through a PersistentVolume and a PersistentVolumeClaim.

> When using the `file` storage backend you must pre-create the PersistentVolume and PersistentVolumeClaim the LeanIX Kubernetes Connector should use.

The `azureblob` storage backend leverages an Azure Storage account you must provide to store the `.ldif` and `.log` files.

### Installation - Helm chart

Before you can install the LeanIX Kubernetes Connector make sure that the following pre-requisites are fulfilled on your local workstation.

- [kubectl is installed](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [helm client is installed](https://helm.sh/docs/using_helm/#installing-the-helm-client)
- [git is installed](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

#### **Pre-release versions**

If you want to install pre-release versions of the LeanIX Kubernetes Connector, append the `helm` CLI commands with `--devel`.

#### **Helm 2 requirements**

On the server-side the Helm server component Tiller must be deployed into the Kubernetes cluster, when you are not using Helm 3.

- [Installing Tiller](https://helm.sh/docs/using_helm/#installing-tiller)

#### **Add LeanIX Kubernetes Connector Helm chart repository**

Before you can install the LeanIX Kubernetes Connector via the provided Helm chart you must add the Helm chart repository first.

``` bash
helm repo add leanix 'https://raw.githubusercontent.com/leanix/leanix-k8s-connector/master/helm/'
helm repo update
helm repo list
```

The output of the `helm repo list` command should look like this.

``` bash
NAME                  URL
stable                https://kubernetes-charts.storage.googleapis.com
local                 http://127.0.0.1:8879/charts
leanix                https://raw.githubusercontent.com/leanix/leanix-k8s-connector/master/helm/
```

A list of the available LeanIX Kubernetes connector Helm chart versions can be retrieved with the command `helm search repo leanix`.

``` bash
NAME                                        CHART VERSION APP VERSION DESCRIPTION
leanix/leanix-k8s-connector                 1.0.0         1.0.0       Retrieves information from Kubernetes cluster
```

#### Setting up data source in LeanIX workspace

Create a new data source with k8s connector template(mentioned above) and add required parameters

##### Integration Hub connector configuration

|      Name       |     Format      | Default | Mandatory |
| :-------------: | :-------------: | :-----: | :-------: |
| resolveStrategy | label/namespace |  empty  |    ✅     |
|  resolveLabel   |   plain text    |  empty  |    ❌     |


#### **Starting Connector in k8s**

> **_NOTE:_** The LeanIX Integration API options requires an API token. See the LeanIX technical documentation on how to obtain one. [LeanIX Technical Documentation](https://dev.leanix.net/docs/authentication#section-generate-api-tokens)

> **_NOTE:_** Make sure Integration Hub data source is setup on the workspace

Create a Kubernetes secret with the LeanIX API token.

``` bash
kubectl create secret generic api-token --from-literal=token={LEANIX_API_TOKEN}
```

The following configuration example assumes that you use the `azureblob` storage backend.

| Parameter                 | Default value | Provided value                       | Notes                                                                                                                                                                                                                                  |
| ------------------------- | ------------- | ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| integrationApi.fqdn       | ""            | app.leanix.net                       | The FQDN of your LeanIX instance                                                                                                                                                                                                       |
| integrationApi.secretName | ""            | api-token                            | The name of the Kubernetes secret containing the LeanIX API token.                                                                                                                                                                     |
| integrationApi.datasourceName | ""        | aks-cluster-k8s-connector            | The name of the datasource configured on the workspace
| schedule.standard         | 0 */1 * * *   |                                      | CronJob schedule. Defaults to every hour, when you enabled the LeanIX Integration API option. Schedule lowest possible value is every hour                                                                                             |
| clustername               | kubernetes    | aks-cluster                          | The name of the Kubernetes cluster.                                                                                                                                                                                                    |
| connectorID               | Random UUID   | aks-cluster                          | The name of the Kubernetes cluster. If not provided a random UUID is generated per default.                                                                                                                                            |
| connectorVersion          | "1.0.0"       | "1.0.0"                              | The version that is used in the LeanIX Integration API processor configuration. Defaults to 1.0.0.                                                                                                                                     |
| processingMode            | "full"        | "full"                               | The processing mode of the LeanIX Integration API processor configuration. Defaults to partial.                                                                                                                                        |
| lxWorkspace               | ""            | 00000000-0000-0000-0000-000000000000 | The UUID of the LeanIX workspace the data is sent to. Make sure Integration Hub data source is also setup in the same workspace                                                                                                        |
| verbose                   | false         | true                                 | Enables verbose logging on the stdout interface of the container.                                                                                                                                                                      |
| storageBackend            | file          | azureblob                            | The default value for the storage backend is `file`, if not provided.                                                                                                                                                                  |
| secretName                | ""            | azure-secret                         | The name of the Kubernetes secret containing the Azure Storage account credentials.                                                                                                                                                    |
| container                 | ""            | leanixk8sconnector                   | The name of the container used to store the `kubernetes.ldif` and `leanix-k8s-connector.log` files.                                                                                                                                    |
| blacklistNameSpaces       | kube-system   | kube-system, default                 | Namespaces that are not scanned by the connector. Must be provided in the format `"{kube-system,default}"` when using the `--set` option. Wildcard blacklisting is also supported e.g. `"{kube-*,default}"` or `"{*-system,default}"`. |

``` bash
helm upgrade --install leanix-k8s-connector leanix/leanix-k8s-connector \
--set integrationApi.fqdn=app.leanix.net \
--set integrationApi.secretName=api-token \
--set integrationApi.datasourceName=aks-cluster-k8s-connector \
--set args.clustername=aks-cluster \
--set args.connectorID=aks-cluster \
--set args.connectorVersion=1.0.0 \
--set args.processingMode=full \
--set args.lxWorkspace=00000000-0000-0000-0000-000000000000 \
--set args.verbose=true \
--set args.storageBackend=azureblob \
--set args.azureblob.secretName=azure-secret \
--set args.azureblob.container=leanixk8sconnector \
--set args.blacklistNamespaces="{kube-system,default}"
```

Beside the option to override the default values and provide values via the `--set` option of the `helm` command, you can also edit the `values.yaml` file.

``` yaml
...
integrationApi:
  fqdn: "app.leanix.net"
  secretName: "api-token"
  datasourceName: "aks-cluster-k8s-connector"

schedule:
  standard: "0 */1 * * *"
...
args:
  clustername: aks-cluster
  connectorID: aks-cluster
  connectorVersion: "1.0.0"
  processingMode: full
  lxWorkspace: "00000000-0000-0000-0000-000000000000"
  verbose: true
  storageBackend: azureblob
  file:
    localFilePath: "/mnt/leanix-k8s-connector"
    claimName: ""
  azureblob:
    secretName: "azure-secret"
    container: "leanixk8sconnector"
  blacklistNamespaces:
  - "kube-system"
  - "default"
...
```

#### **file storage backend**

The first step to get started with the `file` storage backend type is to create the PersistentVolume and PersistentVolumeClaim in advance.

In the following example the creation of a PV and PVC to connect to Azure Files is shown.

Start with the creation of an Azure Storage account and an Azure file share as described in the Azure documentation. In our example we used _leanixk8sconnector_ as file share name.

1. [Create a storage account](https://docs.microsoft.com/en-us/azure/storage/common/storage-quickstart-create-account?tabs=azure-portal)
2. [Create a file share in Azure Files](https://docs.microsoft.com/en-us/azure/storage/files/storage-how-to-create-file-share)

Next, create a Kubernetes secret with the Azure Storage account name and the Azure Storage account key. The information about the name and the key can be retrieved directly via the Azure portal.

- [Access keys](https://docs.microsoft.com/en-us/azure/storage/common/storage-account-manage#access-keys)

``` bash
kubectl create secret generic azure-secret --from-literal=azurestorageaccountname={STORAGE_ACCOUNT_NAME} --from-literal=azurestorageaccountkey={STORAGE_KEY}
```

Afterwards create the PV and PVC using the template below running the `kubectl apply -f template.yaml` command.

``` yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: azurefile
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteMany
  persistentVolumeReclaimPolicy: Retain
  azureFile:
    secretName: azure-secret
    shareName: leanixk8sconnector
    readOnly: false
  mountOptions:
  - dir_mode=0777
  - file_mode=0777
  - uid=1000
  - gid=1000
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: azurefile
spec:
  accessModes:
    - ReadWriteMany
  volumeName: azurefile
  storageClassName: ""
  resources:
    requests:
      storage: 1Gi
```

Run `kubectl get pv && kubectl get pvc` and check your output. It should look like this.

``` bash
NAME        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM               STORAGECLASS   REASON   AGE
azurefile   1Gi        RWX            Retain           Bound    default/azurefile                           45s
NAME        STATUS   VOLUME      CAPACITY   ACCESS MODES   STORAGECLASS   AGE
azurefile   Bound    azurefile   1Gi        RWX                           45s
```

Finally, we use the Helm chart deploying the LeanIX Kubernetes Connector to the Kubernetes cluster.

The following command deploys the connector to the Kubernetes cluster and overwrites the parameters in the `values.yaml` file.

| Parameter           | Default value             | Provided value                       | Notes |
| ------------------- | ------------------------- | ------------------------------------ | ----- |
| integrationApi.fqdn       | ""            | app.leanix.net                       | The FQDN of your LeanIX instance                                                                                                                                                                                                       |
| integrationApi.secretName | ""            | api-token                            | The name of the Kubernetes secret containing the LeanIX API token.                                                                                                                                                                     |
| integrationApi.datasourceName | ""        | aks-cluster-k8s-connector            | The name of the datasource configured on the workspace
| schedule.standard   | 0 */1 * * *               |                                      | CronJob schedule. Defaults to every hour. |
| clustername         | kubernetes                | aks-cluster                          | The name of the Kubernetes cluster. |
| connectorID         | Random UUID               | aks-cluster                          | The name of the Kubernetes cluster. If not provided a random UUID is generated per default. |
| connectorVersion    | "1.0.0"                   | "1.0.0"                              | The version that is used in the LeanIX Integration API processor configuration. Defaults to 1.0.0. |
| processingMode      | "full"                    | "full"                               | The processing mode of the LeanIX Integration API processor configuration. Defaults to partial. |
| lxWorkspace         | ""                        | 00000000-0000-0000-0000-000000000000 | The UUID of the LeanIX workspace the data is sent to. |
| verbose             | false                     | true                                 | Enables verbose logging on the stdout interface of the container. |
| storageBackend      | file                      |                                      | The default value for the storage backend is `file`, if not provided. |
| localFilePath       | /mnt/leanix-k8s-connector |                                      | The path that is used for mounting the PVC into the container and storing the `kubernetes.ldif` and `leanix-k8s-connector.log` files. |
| claimName           | ""                        | azurefile                            | The name of the PVC used to store the `kubernetes.ldif` and `leanix-k8s-connector.log` files. |
| blacklistNameSpaces | kube-system               | kube-system, default                 | Namespaces that are not scanned by the connector. Must be provided in the format `"{kube-system,default}"` when using the `--set` option. Wildcard blacklisting is also supported e.g. `"{kube-*,default}"` or `"{*-system,default}"`. |

``` bash
helm upgrade --install leanix-k8s-connector leanix/leanix-k8s-connector \
--set integrationApi.fqdn=app.leanix.net \
--set integrationApi.secretName=api-token \
--set integrationApi.datasourceName=aks-cluster-k8s-connector \
--set args.clustername=aks-cluster \
--set args.connectorID=aks-cluster \
--set args.connectorVersion=1.0.0 \
--set args.processingMode=full \
--set args.lxWorkspace=00000000-0000-0000-0000-000000000000 \
--set args.verbose=true \
--set args.file.claimName=azurefile \
--set args.blacklistNamespaces="{kube-system,default}"
```

Beside the option to override the default values and provide values via the `--set` option of the `helm` command, you can also edit the `values.yaml` file.

``` yaml
...
integrationApi:
  fqdn: "app.leanix.net"
  secretName: "api-token"
  datasourceName: "aks-cluster-k8s-connector"

schedule:
  standard: "0 */1 * * *"
...
args:
  clustername: aks-cluster
  connectorID: aks-cluster
  connectorVersion: "1.0.0"
  processingMode: full
  lxWorkspace: "00000000-0000-0000-0000-000000000000"
  verbose: true
  storageBackend: file
  file:
    localFilePath: "/mnt/leanix-k8s-connector"
    claimName: "azurefile"
  azureblob:
    secretName: ""
    container: ""
  blacklistNamespaces:
  - "kube-system"
  - "default"
...
```

#### **azureblob storage backend**

The first step to get started with the `azureblob` storage backend type is to create an Azure Storage account as described in the Azure documentation.

1. [Create a storage account](https://docs.microsoft.com/en-us/azure/storage/common/storage-quickstart-create-account?tabs=azure-portal)

Next, create a Kubernetes secret which contains the Azure Storage account name and the Azure Storage account key. The information about the name and the key can be retrieved directly via the Azure portal.

- [Access keys](https://docs.microsoft.com/en-us/azure/storage/common/storage-account-manage#access-keys)

``` bash
kubectl create secret generic azure-secret --from-literal=azurestorageaccountname={STORAGE_ACCOUNT_NAME} --from-literal=azurestorageaccountkey={STORAGE_KEY}
```

Finally, we use the Helm chart deploying the LeanIX Kubernetes Connector to the Kubernetes cluster.

The following command deploys the connector to the Kubernetes cluster and overwrites the parameters in the `values.yaml` file.

| Parameter           | Default value | Provided value                       | Notes |
| ------------------- | ------------- | ------------------------------------ | ----- |
| integrationApi.fqdn       | ""            | app.leanix.net                       | The FQDN of your LeanIX instance                                                                                                                                                                                                       |
| integrationApi.secretName | ""            | api-token                            | The name of the Kubernetes secret containing the LeanIX API token.                                                                                                                                                                     |
| integrationApi.datasourceName | ""        | aks-cluster-k8s-connector            | The name of the datasource configured on the workspace
| schedule.standard   | 0 */1 * * *               |                                      | CronJob schedule. Defaults to every hour. |
| clustername         | kubernetes    | aks-cluster                          | The name of the Kubernetes cluster. |
| connectorID         | Random UUID   | aks-cluster                          | The name of the Kubernetes cluster. If not provided a random UUID is generated per default. |
| connectorVersion    | "1.0.0"       | "1.0.0"                              | The version that is used in the LeanIX Integration API processor configuration. Defaults to 1.0.0. |
| processingMode      | "full"        | "full"                               | The processing mode of the LeanIX Integration API processor configuration. Defaults to partial. |
| lxWorkspace         | ""            | 00000000-0000-0000-0000-000000000000 | The UUID of the LeanIX workspace the data is sent to. |
| verbose             | false         | true                                 | Enables verbose logging on the stdout interface of the container. |
| storageBackend      | file          | azureblob                            | The default value for the storage backend is `file`, if not provided. |
| secretName          | ""            | azure-secret                         | The name of the Kubernetes secret containing the Azure Storage account credentials. |
| container           | ""            | leanixk8sconnector                   | The name of the container used to store the `kubernetes.ldif` and `leanix-k8s-connector.log` files. |
| blacklistNameSpaces | kube-system   | kube-system, default                 | Namespaces that are not scanned by the connector. Must be provided in the format `"{kube-system,default}"` when using the `--set` option. Wildcard blacklisting is also supported e.g. `"{kube-*,default}"` or `"{*-system,default}"`. |

``` bash
helm upgrade --install leanix-k8s-connector leanix/leanix-k8s-connector \
--set integrationApi.fqdn=app.leanix.net \
--set integrationApi.secretName=api-token \
--set integrationApi.datasourceName=aks-cluster-k8s-connector \
--set args.clustername=aks-cluster \
--set args.connectorID=aks-cluster \
--set args.connectorVersion=1.0.0 \
--set args.processingMode=full \
--set args.lxWorkspace=00000000-0000-0000-0000-000000000000 \
--set args.verbose=true \
--set args.storageBackend=azureblob \
--set args.azureblob.secretName=azure-secret \
--set args.azureblob.container=leanixk8sconnector \
--set args.blacklistNamespaces="{kube-system,default}"
```

Beside the option to override the default values and provide values via the `--set` option of the `helm` command, you can also edit the `values.yaml` file.

``` yaml
...
integrationApi:
  fqdn: "app.leanix.net"
  secretName: "api-token"
  datasourceName: "aks-cluster-k8s-connector"

schedule:
  standard: "*/1 * * * *"
  integrationApi: "0 */1 * * *"
...
args:
  clustername: aks-cluster
  connectorID: aks-cluster
  connectorVersion: "1.0.0"
  processingMode: full
  lxWorkspace: "00000000-0000-0000-0000-000000000000"
  verbose: true
  storageBackend: azureblob
  file:
    localFilePath: "/mnt/leanix-k8s-connector"
    claimName: ""
  azureblob:
    secretName: "azure-secret"
    container: "leanixk8sconnector"
  blacklistNamespaces:
  - "kube-system"
  - "default"
...
```

#### **Optional - Advanced deployment settings**

Depending on your corporate policies / permission set the creation of ClusterRoles or ClusterRoleBindings are done beforehand. You then can set in the `values.yaml` the following setting to `true` or use `--set` when installing the Helm chart to override the default value.

``` yaml
...
clusterRoleAlreadyCreated: true
...
```

Furthermore, when you use different user ids and group ids in your environment than the provided default. You can set them in the `values.yaml` or use `--set` when installing the Helm chart to override the default values.

``` yaml
...
securityContext:
  userId: 1337
  groupId: 1337
...
```

If you are in need to provide additional ENV values you can do so by setting them in the `values.yaml` or use `--set` when installing the Helm chart to override the default values.

``` yaml
...
args:
...
  additionalEnv:
    FOO: "BAR"
...
```

### Developer Environment Setup
> **_NOTE:_** Make sure Integration Hub data source is setup on the workspace
 
The connector can be published to a minikube instance

## Steps
1. install minikube
2. If already installed make sure you run
    > `minikube delete`
3. Start minikube instance
   > `minikube start --insecure-registry="<your-ip eg:192. ..>:5000"`
4. Open minikube dashboard
   >  `minikube dashboard`

By default, The cronJob pulls the image from docker hub. To override the behaviour:

1. Install local docker image registry (https://www.docker.com/blog/how-to-use-your-own-registry-2/)
2. Run registry locally
   > `docker run -d -p 5000:5000 --name registry registry:2.7`
3. Use Makefile.local
    > `
        make -f Makefile.local clean build image
        docker tag leanix-dev/leanix-k8s-connector localhost:5000/leanix-dev/leanix-k8s-connector:1.0.0-dev
        docker push localhost:5000/leanix-dev/leanix-k8s-connector:1.0.0-dev
    `
   
   - or `make -f Makefile.local local`
4. Finally, run the command

    > `
    helm upgrade --install --devel leanix-k8s-connector leanix/leanix-k8s-connector \
    --set image.repository=192.168.29.244:5000/leanix-dev/leanix-k8s-connector  \
    --set image.tag=1.0.0-dev \
    --set args.clustername=minikube \
    --set args.connectorID=minikube \
    --set integrationApi.fqdn=app.leanix.net \
    --set integrationApi.secretName=api-token \
    --set integrationApi.datasourceName=k8s-connector-test \
    --set args.lxWorkspace=00000000-0000-0000-0000-000000000000 \
    --set args.verbose=true \
    --set args.storageBackend=azureblob \
    --set args.azureblob.secretName=azure-secret \
    --set args.azureblob.container=leanixk8sconnector \
    --set args.blacklistNamespaces="{kube-system}"
    `

Make sure you are using `--devel` flag to get the latest helm chart version

## Known issues

If the LeanIX Kubernetes Connector pod resides in an `Error` or `CrashLoopBackOff` state and you issued a `helm upgrade --install` command to fix it, you still the see the same pod instead of a new one.

This is not an issue of the LeanIX Kubernetes Connector itself. Instead it takes several minutes in this case until the `CronJob` creates a new pod.

If you do not want to wait until Kubernetes fix it itself, you can just delete the `Job` object.

Run `kubectl get jobs.batch` and look for the `Job` object with COMPLETIONS 0/1.

``` bash
NAME                              COMPLETIONS   DURATION   AGE
leanix-k8s-connector-1563961200   0/1           20m        20m
```

Issue `kubectl delete jobs.batch leanix-k8s-connector-1563961200` and you should see a new pod coming up afterwards.

## Migration docs

### 3.0.0 to 4.0.0
- Converted to a self-start connector of Integration Hub. Data source must be configured in the workspace before setting up the connector.
- New **mandatory** flag is introduced to work with Integration hub data source - `integrationApi.datasourceName`
- LDIF is still uploaded to choosen backend including the Integration Hub to trigger Integration API automatically. Hence `integrationApi.enabled` flag is removed
- All the flags which are required when `integrationapi.enabled` is true should be passed such as `integrationApi.fqdn`, `integrationApi.secretName`
- Integration API connector is automatically provisioned to the workspace. Dependency on cloud-beta is removed by introducing custom fields - `resolveStrategy`, `resolveLabel`
- Integration API connector type is changed to `leanix-mi-connector` and connector id to `leanix-k8s-connector` hence the connector version is changed to `1.0.0`. The default value is also changed to `1.0.0` from `1.1.1`
- `schedule.integrationApi` flag is removed and there is a single `schedule.standard`
- Lowest possible value for `schedule.standard` is every hour

## Version history

[CHANGELOG](CHANGELOG.md)

| Release date | Connector version | Integration version | Helm chart version | Container image tag |
| :----------: | :---------------: | :-----------------: | :----------------: | :-----------------: |
|  2021-09-16  |       5.0.0       |        1.0.0        |       5.0.0        |        5.0.0        |
|  2021-08-30  |       4.0.0       |        1.0.0        |       4.0.0        |        4.0.0        |
|  2021-08-04  |       3.0.0       |        1.0.0        |       3.0.0        |        3.0.0        |
|  2021-06-21  |       2.0.0       |        1.0.0        |       2.0.0        |        2.0.0        |
|  2020-10-22  |    2.0.0-beta7    |        1.0.0        |    2.0.0-beta7     |     2.0.0-beta7     |
|  2020-10-21  |    2.0.0-beta6    |        1.0.0        |    2.0.0-beta6     |     2.0.0-beta6     |
|  2020-10-15  |    2.0.0-beta5    |        1.0.0        |    2.0.0-beta5     |     2.0.0-beta5     |
|  2020-06-15  |    2.0.0-beta4    |        1.0.0        |    2.0.0-beta4     | 2.0.0-beta4-1b46be5 |
|  2020-04-28  |    2.0.0-beta3    |        1.0.0        |    2.0.0-beta3     | 2.0.0-beta3-fa5ea6f |
|  2020-02-07  |    2.0.0-beta2    |        1.0.0        |    2.0.0-beta1     | 2.0.0-beta2-f8218f4 |
|  2020-01-14  |    2.0.0-beta1    |        1.0.0        |    2.0.0-beta1     | 2.0.0-beta1-d5555d2 |
|  2019-09-26  |       1.1.0       |        1.0.0        |       1.0.0        |       23d019b       |
|  2019-08-28  |       1.0.0       |        1.0.0        |       1.0.0        |       b0bc069       |
