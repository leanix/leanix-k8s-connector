package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/leanix"
	"github.com/leanix/leanix-k8s-connector/pkg/mapper"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"github.com/leanix/leanix-k8s-connector/pkg/version"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/op/go-logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	clusterNameFlag                  string = "clustername"
	storageBackendFlag               string = "storage-backend"
	azureAccountNameFlag             string = "azure-account-name"
	azureAccountKeyFlag              string = "azure-account-key"
	azureContainerFlag               string = "azure-container"
	localFilePathFlag                string = "local-file-path"
	verboseFlag                      string = "verbose"
	connectorIDFlag                  string = "connector-id"
	connectorVersionFlag             string = "connector-version"
	connectorProcessingModeFlag      string = "processing-mode"
	integrationAPIDatasourceNameFlag string = "integration-api-datasourcename"
	integrationAPIFqdnFlag           string = "integration-api-fqdn"
	integrationAPITokenFlag          string = "integration-api-token"
	blacklistNamespacesFlag          string = "blacklist-namespaces"
	lxWorkspaceFlag                  string = "lx-workspace"
	localFlag                        string = "local"
)

const (
	lxVersion                      string = "1.0.0"
	lxConnectorID                  string = "Kubernetes"
	lxConnectorType                string = "leanix-k8s-connector"
	lxConnectorProcessingDirection string = "inbound"
)

var log = logging.MustGetLogger("leanix-k8s-connector")

func main() {
	stdoutLogger, debugLogBuffer := initLogger()
	err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}
	enableVerbose(stdoutLogger, viper.GetBool(verboseFlag))

	log.Info("----------Attempting to Self Start via Integration Hub----------")

	accessToken, err := leanix.Authenticate(viper.GetString(integrationAPIFqdnFlag), viper.GetString(integrationAPITokenFlag))
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Integration Hub authentication successful.")
	startResponse, err := leanix.SelfStartRun(viper.GetString(integrationAPIFqdnFlag), accessToken, viper.GetString(integrationAPIDatasourceNameFlag))
	if err != nil {
		log.Info("Failed to start Integration Hub. Terminating..")
		log.Fatal(err)
		return
	}
	if startResponse != nil {
		log.Infof("Successfully self started via Integration Hub. Progress call back - %s", startResponse.ProgressCallbackUrl)
	}

	log.Info("----------Start----------")
	log.Infof("LeanIX Kubernetes connector build version: %s", version.VERSION)
	log.Infof("LeanIX integration version: %s", lxVersion)
	log.Infof("LeanIX connector id: %s", lxConnectorID)
	log.Infof("LeanIX connector type: %s", lxConnectorType)
	log.Infof("LeanIX connector version: %s", viper.GetString(connectorVersionFlag))
	log.Infof("LeanIX connector processing direction: %s", lxConnectorProcessingDirection)
	log.Infof("LeanIX connector processing mode: %s", viper.GetString(connectorProcessingModeFlag))
	log.Infof("Target LeanIX workspace: %s", viper.GetString(lxWorkspaceFlag))
	log.Infof("Target Kubernetes cluster name: %s", viper.GetString(clusterNameFlag))

	var config *restclient.Config
	if viper.GetBool(localFlag) {
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// use the current context in kubeconfig
		config, err = restclient.InClusterConfig()
		if err != nil {
			log.Fatalf("Failed to load kube config. Running in Kubernetes?\n%s", err)
		}
	}

	log.Debugf("Kubernetes master from config: %s", config.Host)

	kubernetesAPI, err := kubernetes.NewAPI(config)
	if err != nil {
		log.Fatal(err)
	}
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("Get blacklist namespaces list...")
	blacklistedNamespacesList := viper.GetStringSlice(blacklistNamespacesFlag)
	blacklistedNamespaces, err := kubernetesAPI.Namespaces(blacklistedNamespacesList)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Getting blacklist namespaces list done.")
	log.Infof("Namespace blacklist: %v", reflect.ValueOf(blacklistedNamespaces).MapKeys())

	resourcesList, err := ServerPreferredListableResources(kubernetesAPI.Client.Discovery())
	if err != nil {
		log.Fatal(err)
	}
	groupVersionResources, err := discovery.GroupVersionResources(resourcesList)
	if err != nil {
		log.Panic(err)
	}

	log.Debug("Listing nodes...")
	nodes, err := kubernetesAPI.Nodes()
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Listing nodes done.")

	log.Debug("Map nodes to Kubernetes object")
	clusterKubernetesObject, err := mapper.MapNodes(
		viper.GetString("clustername"),
		nodes,
	)
	if err != nil {
		log.Fatal(err)
	}

	kubernetesObjects := make([]mapper.KubernetesObject, 0)
	kubernetesObjects = append(kubernetesObjects, *clusterKubernetesObject)

	resourceGroupWhitelist := map[string]map[string]interface{}{
		"": map[string]interface{}{
			"serviceaccounts":        struct{}{},
			"services":               struct{}{},
			"nodes":                  struct{}{},
			"pods":                   struct{}{},
			"namespaces":             struct{}{},
			"configmaps":             struct{}{},
			"persistentvolumes":      struct{}{},
			"persistentvolumeclaims": struct{}{},
			"replicationcontrollers": struct{}{},
		},
		"apps": map[string]interface{}{
			"deployments":  struct{}{},
			"statefulsets": struct{}{},
			"daemonsets":   struct{}{},
			"replicasets":  struct{}{},
		},
		"apiextensions.k8s.io": map[string]interface{}{
			"customresourcedefinitions": struct{}{},
		},
		"rbac.authorization.k8s.io": map[string]interface{}{
			"clusterrolebindings": struct{}{},
			"rolebindings":        struct{}{},
			"clusterroles":        struct{}{},
			"roles":               struct{}{},
		},
		"networking.k8s.io": map[string]interface{}{
			"ingresses":       struct{}{},
			"networkpolicies": struct{}{},
		},
		"autoscaling": map[string]interface{}{
			"horizontalpodautoscalers": struct{}{},
		},
		"policy": map[string]interface{}{
			"podsecuritypolicies": struct{}{},
		},
		"storage.k8s.io": map[string]interface{}{
			"storageclasses": struct{}{},
		},
		"batch": map[string]interface{}{
			"cronjobs": struct{}{},
			"jobs":     struct{}{},
		},
	}

	for gvr := range groupVersionResources {
		if _, ok := resourceGroupWhitelist[gvr.Group][gvr.Resource]; !ok {
			log.Debugf("Not scanning resouce %s", strings.Join([]string{gvr.Group, gvr.Version, gvr.Resource}, "/"))
			continue
		}
		instances, err := dynClient.Resource(gvr).List(metav1.ListOptions{})
		if err != nil {
			log.Panic(err)
		}
		for _, i := range instances.Items {
			if _, ok := blacklistedNamespaces[i.GetNamespace()]; ok {
				continue
			}
			nko := mapper.KubernetesObject{
				Type: i.GetKind(),
				ID:   string(i.GetUID()),
				Data: i.Object,
			}
			kubernetesObjects = append(kubernetesObjects, nko)
		}
	}

	customFields := mapper.CustomFields{
		ConnectorInstance: viper.GetString(connectorIDFlag),
		BuildVersion:      version.VERSION,
	}

	ldif := mapper.LDIF{
		ConnectorID:         lxConnectorID,
		ConnectorType:       lxConnectorType,
		ConnectorVersion:    viper.GetString(connectorVersionFlag),
		ProcessingDirection: lxConnectorProcessingDirection,
		ProcessingMode:      viper.GetString(connectorProcessingModeFlag),
		LxVersion:           lxVersion,
		LxWorkspace:         viper.GetString(lxWorkspaceFlag),
		Description:         "Map Kubernetes objects to LeanIX Fact Sheets",
		CustomFields:        customFields,
		Content:             kubernetesObjects,
	}

	_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Successfully requested data. Uploaded ldif to configured "+viper.GetString("storage-backend"))
	if err != nil {
		log.Infof("Failed to progress[%s] to Integration Hub", leanix.INPROGRESS)
	}
	log.Debug("Marshal ldif")
	ldifByte, err := storage.Marshal(ldif)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Upload %s to %s", storage.LdifFileName, viper.GetString("storage-backend"))
	azureOpts := storage.AzureBlobOpts{
		AccountName: viper.GetString(azureAccountNameFlag),
		AccountKey:  viper.GetString(azureAccountKeyFlag),
		Container:   viper.GetString(azureContainerFlag),
	}
	localFileOpts := storage.LocalFileOpts{
		Path: viper.GetString(localFilePathFlag),
	}
	uploader, err := storage.NewBackend(viper.GetString("storage-backend"), &azureOpts, &localFileOpts)
	if err != nil {
		log.Fatal(err)
	}
	err = uploader.UploadLdif(ldifByte)
	if err != nil {
		log.Fatal(err)
		_, err := leanix.UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "Failed to upload ldif to backend storage")
		if err != nil {
			return
		}
	}
	_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Successfully uploaded ldif to backend: "+viper.GetString("storage-backend"))
	if err != nil {
		log.Infof("Failed to update progress[%s] to Integration Hub", leanix.INPROGRESS)
	}
	_, err = leanix.UploadLdif(startResponse.LdifResultUrl, ldifByte)
	if err != nil {
		log.Debug("Failed to upload ldif to Integration Hub ldif SAS Url")
		_, err := leanix.UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "Failed to upload ldif to Integration Hub ldif SAS Url")
		log.Fatal(err)
		return
	}
	_, err = leanix.UpdateProgress(startResponse.ProgressCallbackUrl, leanix.FINISHED, "")
	if err != nil {
		log.Infof("Failed to progress[%s] to Integration Hub", leanix.FINISHED)
	}
	log.Debug("-----------End-----------")
	err = uploader.UploadLog(debugLogBuffer.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	log.Info("-----------End-----------")
}

func ServerPreferredListableResources(d discovery.DiscoveryInterface) ([]*metav1.APIResourceList, error) {
	all, err := discovery.ServerPreferredResources(d)
	return discovery.FilteredBy(discovery.ResourcePredicateFunc(func(groupVersion string, r *metav1.APIResource) bool {
		return strings.Contains(r.Verbs.String(), "list")
	}), all), err
}

func parseFlags() error {
	flag.String(clusterNameFlag, "", "unique name of the Kubernetes cluster")
	flag.String(storageBackendFlag, storage.FileStorage, fmt.Sprintf("storage where the %s file is placed (%s, %s)", storage.LdifFileName, storage.FileStorage, storage.AzureBlobStorage))
	flag.String(azureAccountNameFlag, "", "Azure storage account name")
	flag.String(azureAccountKeyFlag, "", "Azure storage account key")
	flag.String(azureContainerFlag, "", "Azure storage account container")
	flag.String(localFilePathFlag, ".", "path to place the ldif file when using local file storage backend")
	flag.Bool(verboseFlag, false, "verbose log output")
	flag.String(connectorIDFlag, "", "unique id of the LeanIX Kubernetes connector")
	flag.String(connectorVersionFlag, "1.0.0", "connector version defaults to 1.0.0 if not specified")
	flag.String(connectorProcessingModeFlag, "partial", "processing mode defaults to partial if not specified")
	flag.String(integrationAPIDatasourceNameFlag, "", "LeanIX Integration Hub Datasource name created on the workspace")
	flag.String(integrationAPIFqdnFlag, "app.leanix.net", "LeanIX Instance FQDN")
	flag.String(integrationAPITokenFlag, "", "LeanIX API token")
	flag.StringSlice(blacklistNamespacesFlag, []string{""}, "list of namespaces that are not scanned")
	flag.String(lxWorkspaceFlag, "", "name of the LeanIX workspace the data is sent to")
	flag.Bool(localFlag, false, "use local kubeconfig from home folder")
	flag.Parse()
	// Let flags overwrite configs in viper
	err := viper.BindPFlags(flag.CommandLine)
	if err != nil {
		return err
	}
	// Check for config values in env vars
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	if viper.GetString(clusterNameFlag) == "" {
		return fmt.Errorf("%s flag must be set", clusterNameFlag)
	}
	if viper.GetString(connectorIDFlag) == "" {
		return fmt.Errorf("%s flag must be set", connectorIDFlag)
	}
	if viper.GetString(lxWorkspaceFlag) == "" {
		return fmt.Errorf("%s flag must be set", lxWorkspaceFlag)
	}
	if viper.GetString(storageBackendFlag) == "azureblob" {
		if viper.GetString(azureAccountNameFlag) == "" {
			return fmt.Errorf("%s flag must be set", azureAccountNameFlag)
		}
		if viper.GetString(azureAccountKeyFlag) == "" {
			return fmt.Errorf("%s flag must be set", azureAccountKeyFlag)
		}
		if viper.GetString(azureContainerFlag) == "" {
			return fmt.Errorf("%s flag must be set", azureContainerFlag)
		}
	}
	if viper.GetString(integrationAPIDatasourceNameFlag) == "" {
		return fmt.Errorf("%s flag must be set", integrationAPIDatasourceNameFlag)
	}
	return nil
}

// InitLogger initialise the logger for stdout and log file
func initLogger() (logging.LeveledBackend, *bytes.Buffer) {
	format := logging.MustStringFormatter(`%{time} ▶ [%{level:.4s}] %{message}`)
	logging.SetFormatter(format)

	// stdout logging backend
	stdout := logging.NewLogBackend(os.Stdout, "", 0)
	stdoutLeveled := logging.AddModuleLevel(stdout)

	// file logging backend
	var mem bytes.Buffer
	fileLogger := logging.NewLogBackend(&mem, "", 0)
	logging.SetBackend(fileLogger, stdoutLeveled)
	return stdoutLeveled, &mem
}

func enableVerbose(logger logging.LeveledBackend, verbose bool) {
	if verbose {
		logger.SetLevel(logging.DEBUG, "")
	} else {
		logger.SetLevel(logging.INFO, "")
	}
}
