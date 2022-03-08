package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/leanix/leanix-k8s-connector/pkg/leanix"
	"github.com/leanix/leanix-k8s-connector/pkg/mapper"
	"github.com/leanix/leanix-k8s-connector/pkg/newmapper"
	"github.com/leanix/leanix-k8s-connector/pkg/scanner"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"github.com/leanix/leanix-k8s-connector/pkg/version"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/op/go-logging"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	enableCustomStorageFlag          string = "enable-custom-storage"
	storageBackendFlag               string = "storage-backend"
	azureAccountNameFlag             string = "azure-account-name"
	azureAccountKeyFlag              string = "azure-account-key"
	azureContainerFlag               string = "azure-container"
	localFilePathFlag                string = "local-file-path"
	verboseFlag                      string = "verbose"
	integrationAPIDatasourceNameFlag string = "integration-api-datasourcename"
	integrationAPIFqdnFlag           string = "integration-api-fqdn"
	integrationAPITokenFlag          string = "integration-api-token"
	blacklistNamespacesFlag          string = "blacklist-namespaces"
	lxWorkspaceFlag                  string = "lx-workspace"
	localFlag                        string = "local"
	kubeConfigPathFlag               string = "kube-config-path"
)

var log = logging.MustGetLogger("leanix-k8s-connector")

func main() {
	stdoutLogger, debugLogBuffer := initLogger()
	err := parseFlags()
	if err != nil {
		log.Error(err)
	}
	enableVerbose(stdoutLogger, viper.GetBool(verboseFlag))

	// use the current context in kubeconfig
	startResponse, err := KubernetesScan(debugLogBuffer)
	if err != nil {
		log.Error(err)
	}
	log.Info("Uploading connector logs to iHub")
	err = storage.UploadConnectorLog(startResponse.ConnectorLoggingUrl, debugLogBuffer.Bytes())
	if err != nil {
		log.Error(err)
	}
}

func KubernetesScan(debugLogBuffer *bytes.Buffer) (response *leanix.SelfStartResponse, err error) {
	log.Info("----------Attempting to Self Start via Integration Hub----------")

	accessToken, err := leanix.Authenticate(viper.GetString(integrationAPIFqdnFlag), viper.GetString(integrationAPITokenFlag))
	if err != nil {
		log.Error(err)
		log.Info("Failed to authenticate. Terminating..")
		return nil, err
	}
	log.Info("Integration Hub authentication successful.")
	startResponse, err := leanix.SelfStartRun(viper.GetString(integrationAPIFqdnFlag), accessToken, viper.GetString(integrationAPIDatasourceNameFlag))
	if err != nil {
		log.Info("Main: Failed to start Integration Hub. Terminating..")
		return startResponse, err
	}
	log.Info("Getting connector config...")
	if startResponse != nil {
		log.Infof("Successfully self started via Integration Hub. Progress call back - %s", startResponse.ProgressCallbackUrl)
		_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Successfully self started via Integration Hub. Connector is in progress")
		if err != nil {
			log.Infof("KubernetesScan: Failed to update progress[%s] to Integration Hub", leanix.IN_PROGRESS)
		}
	}

	log.Info("----------Start----------")
	log.Infof("LeanIX Kubernetes connector build version: %s", version.VERSION)
	log.Infof("LeanIX integration version: %s", startResponse.BindingKey.LXVersion)
	log.Infof("LeanIX connector id: %s", startResponse.BindingKey.ConnectorId)
	log.Infof("LeanIX connector type: %s", startResponse.BindingKey.ConnectorType)
	log.Infof("LeanIX connector processing direction: %s", startResponse.BindingKey.ProcessingDirection)
	log.Infof("Target LeanIX workspace: %s", viper.GetString(lxWorkspaceFlag))
	log.Infof("Target Kubernetes cluster name: %s", startResponse.ConnectorConfiguration.ClusterName)

	var config *restclient.Config
	if viper.GetBool(localFlag) {
		if viper.GetString(kubeConfigPathFlag) != "" {
			config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(viper.GetString(kubeConfigPathFlag)))
		} else {
			config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		}
		if err != nil {
			return nil, err
		}
	} else {

		config, err = restclient.InClusterConfig()
		if err != nil {
			log.Errorf("Failed to load kube config. Running in Kubernetes?\n%s", err)
			return nil, err
		}
	}
	log.Debugf("Kubernetes master from config: %s", config.Host)
	kubernetesObjects := make([]mapper.KubernetesObject, 0)
	if startResponse.BindingKey.ConnectorId == "leanix-k8s-v3-connector" {
		log.Info("Using new k8s-v3 mapper")
		mappedObject, err := newmapper.ScanKubernetes(startResponse.ConnectorConfiguration.ClusterName, config)
		if err != nil {
			return nil, err
		}
		kubernetesObjects = append(kubernetesObjects, mappedObject...)

		_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Mapping of kubernetes objects with new version done. Preparing ldif.")
		if err != nil {
			log.Infof("New Kubernetes Scan: After Mapping: Failed to update progress[%s] to Integration Hub", leanix.IN_PROGRESS)
		}
	} else {
		log.Info("Using current k8s mapper")
		blacklistedNamespacesList := viper.GetStringSlice(blacklistNamespacesFlag)
		mappedObject, err := scanner.ScanKubernetesCurrent(startResponse, blacklistedNamespacesList, config, debugLogBuffer)
		if err != nil {
			return nil, err
		}
		_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Mapping of kubernetes objects with old version done. Preparing ldif.")
		if err != nil {
			log.Infof("Current Kubernetes Scan: After Mapping: Failed to update progress[%s] to Integration Hub", leanix.IN_PROGRESS)
		}
		kubernetesObjects = append(kubernetesObjects, mappedObject...)
	}

	customFields := mapper.CustomFields{
		BuildVersion:          version.VERSION,
		ResolveStrategy:       startResponse.ConnectorConfiguration.ResolveStrategy,
		ResolveLabel:          startResponse.ConnectorConfiguration.ResolveLabel,
		ClusterName:           startResponse.ConnectorConfiguration.ClusterName,
		EnabledLabelWhitelist: startResponse.ConnectorConfiguration.EnabledLabelWhitelist,
	}

	ldif := mapper.LDIF{
		ConnectorID:         startResponse.BindingKey.ConnectorId,
		ConnectorType:       startResponse.BindingKey.ConnectorType,
		ProcessingDirection: startResponse.BindingKey.ProcessingDirection,
		ProcessingMode:      startResponse.BindingKey.ProcessingMode,
		LxVersion:           startResponse.BindingKey.LXVersion,
		LxWorkspace:         viper.GetString(lxWorkspaceFlag),
		Description:         "Map Kubernetes objects to LeanIX Fact Sheets",
		CustomFields:        customFields,
		Content:             kubernetesObjects,
	}

	_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Successfully collected required kubernetes data.")
	if err != nil {
		log.Infof("KubernetesScan: After mapping: Failed to progress[%s] to Integration Hub", leanix.IN_PROGRESS)
	}
	log.Debug("Marshal ldif")
	ldifByte, err := storage.Marshal(ldif)
	if err != nil {
		_, err = leanix.UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "Failed to marshal ldif")
		log.Error(err)
		return nil, err
	}

	if viper.GetBool(enableCustomStorageFlag) {
		_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Uploading ldif to configured storage backend - "+viper.GetString("storage-backend"))
		if err != nil {
			log.Infof("Failed to progress[%s] to Integration Hub", leanix.IN_PROGRESS)
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
			_, err = leanix.UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "Failed to create uploader for backend storage")
			return nil, err
		}
		err = uploader.UploadLdif(ldifByte, storage.LdifFileName+viper.GetString(lxWorkspaceFlag)+storage.LdifFileExtension)
		if err != nil {
			log.Error(err)
			_, err := leanix.UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "Failed to upload ldif to backend storage configured storage backend - "+viper.GetString("storage-backend"))
			if err != nil {
				return nil, err
			}
		}

		_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Successfully uploaded ldif to configured storage backend - "+viper.GetString("storage-backend"))
		if err != nil {
			log.Infof("Failed to update progress[%s] to Integration Hub", leanix.IN_PROGRESS)
		}
		log.Info("-----------Done Uploading to custom storage backend-----------")
	} else {
		log.Infof("Skipping uploading LDIF and log file to custom storage backend. reason - flag is disabled")
	}

	_, err = leanix.UploadLdif(startResponse.LdifResultUrl, ldifByte)
	if err != nil {
		log.Debug("Failed to upload ldif to Integration Hub ldif SAS Url")
		_, err := leanix.UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "Failed to upload ldif to Integration Hub ldif SAS Url")
		log.Error(err)
		return nil, err
	}
	_, err = leanix.UpdateProgress(startResponse.ProgressCallbackUrl, leanix.FINISHED, "")
	if err != nil {
		log.Infof("Failed to update progress[%s] to Integration Hub", leanix.FINISHED)
	}
	log.Debug("-----------End-----------")
	log.Info("-----------End-----------")
	return startResponse, err
}

func parseFlags() error {
	flag.Bool(enableCustomStorageFlag, false, "Disable/enable custom storage backend option")
	flag.String(storageBackendFlag, storage.FileStorage, fmt.Sprintf("storage where the %s file is placed (%s, %s)", storage.LdifFileName, storage.FileStorage, storage.AzureBlobStorage))
	flag.String(azureAccountNameFlag, "", "Azure storage account name")
	flag.String(azureAccountKeyFlag, "", "Azure storage account key")
	flag.String(azureContainerFlag, "", "Azure storage account container")
	flag.String(localFilePathFlag, ".", "path to place the ldif file when using local file storage backend")
	flag.Bool(verboseFlag, false, "verbose log output")
	flag.String(integrationAPIDatasourceNameFlag, "", "LeanIX Integration Hub Datasource name created on the workspace")
	flag.String(integrationAPIFqdnFlag, "app.leanix.net", "LeanIX Instance FQDN")
	flag.String(integrationAPITokenFlag, "", "LeanIX API token")
	flag.StringSlice(blacklistNamespacesFlag, []string{""}, "list of namespaces that are not scanned")
	flag.String(lxWorkspaceFlag, "", "name of the LeanIX workspace the data is sent to")
	flag.Bool(localFlag, false, "use local kubeconfig from home folder")
	flag.String(kubeConfigPathFlag, "", "path to kubeconfig file")
	flag.Parse()
	// Let flags overwrite configs in viper
	err := viper.BindPFlags(flag.CommandLine)
	if err != nil {
		log.Error(err)
		return err
	}
	// Check for config values in env vars
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	if viper.GetString(lxWorkspaceFlag) == "" {
		return fmt.Errorf("%s flag must be set", lxWorkspaceFlag)
	}
	if viper.GetBool(enableCustomStorageFlag) {
		if viper.GetString(storageBackendFlag) == "none" || viper.GetString(storageBackendFlag) == "" {
			return fmt.Errorf("%s flag must be set since %s is enabled", storageBackendFlag, enableCustomStorageFlag)
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
