package main

import (
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	"path/filepath"
	"strings"

	"github.com/leanix/leanix-k8s-connector/pkg/ihub"
	"github.com/leanix/leanix-k8s-connector/pkg/iris"
	"github.com/leanix/leanix-k8s-connector/pkg/leanix"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"github.com/leanix/leanix-k8s-connector/pkg/utils"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	debugLogBuffer := logger.Init()
	err := parseFlags()
	if err != nil {
		logger.Error("Error occurred when parsing flags.", err)
	}

	var config *restclient.Config
	if viper.GetBool(utils.LocalFlag) {
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		if err != nil {
			logger.Error("Failed to load kube config. Running locally?\n%s", err)
		}
	} else {
		config, err = restclient.InClusterConfig()
		if err != nil {
			logger.Error("Failed to load kube config. Running in Kubernetes?\n%s", err)
		}
	}
	accessToken, err := leanix.Authenticate(viper.GetString(utils.IntegrationAPIFqdnFlag), viper.GetString(utils.IntegrationAPITokenFlag))
	if err != nil {
		logger.Error("Error occurred when authenticating.", err)
		logger.Info("Failed to authenticate. Terminating..")
	}
	if viper.GetBool(utils.IrisFlag) {
		logger.Info("Enabled Iris")
		runId := iris.GenerateRunId()
		irisScanner := iris.NewScanner(
			"Iris Integration",
			viper.GetString(utils.IntegrationAPIFqdnFlag), runId,
		)

		err = irisScanner.Scan(config, viper.GetString(utils.LxWorkspaceFlag), viper.GetString(utils.ConfigurationNameFlag), accessToken)
		if err != nil {
			logger.Error("Failed to scan Kubernetes via vsm-iris.\n%s", err)
		}
	} else {
		// use the current context in kubeconfig
		startResponse, err := ihub.KubernetesScan(config, viper.GetViper())
		if err != nil {
			logger.Error("Error occurred when processing kubeconfig", err)
		}
		if startResponse != nil {
			logger.Info("Uploading connector logs to iHub")
			err := logger.Sync()
			if err != nil {
				logger.Error("Failed to sync buffer.", err)
				return
			}
			err = storage.UploadConnectorLog(startResponse.ConnectorLoggingUrl, debugLogBuffer.Bytes())
			if err != nil {
				logger.Error("Error occurred uploading ConnectorLogs", err)
			}
		} else {
			logger.Error("Invalid response from integration hub. Can't upload logs.", err)
		}
	}
}

func parseFlags() error {
	flag.Bool(utils.EnableCustomStorageFlag, false, "Disable/enable custom storage backend option")
	flag.String(utils.StorageBackendFlag, storage.FileStorage, fmt.Sprintf("storage where the %s file is placed (%s, %s)", storage.LdifFileName, storage.FileStorage, storage.AzureBlobStorage))
	flag.String(utils.AzureAccountNameFlag, "", "Azure storage account name")
	flag.String(utils.AzureAccountKeyFlag, "", "Azure storage account key")
	flag.String(utils.AzureContainerFlag, "", "Azure storage account container")
	flag.String(utils.LocalFilePathFlag, ".", "path to place the ldif file when using local file storage backend")
	flag.Bool(utils.VerboseFlag, false, "verbose log output")
	flag.String(utils.IntegrationAPIDatasourceNameFlag, "", "LeanIX Integration Hub Datasource name created on the workspace")
	flag.String(utils.IntegrationAPIFqdnFlag, "app.leanix.net", "LeanIX Instance FQDN")
	flag.String(utils.IntegrationAPITokenFlag, "", "LeanIX API token")
	flag.StringSlice(utils.BlacklistNamespacesFlag, []string{""}, "list of namespaces that are not scanned")
	flag.String(utils.LxWorkspaceFlag, "", "name of the LeanIX workspace the data is sent to")
	flag.Bool(utils.LocalFlag, false, "use local kubeconfig from home folder")
	flag.Bool(utils.IrisFlag, false, "send kubernetes events to new integration api")
	flag.String(utils.ConfigurationNameFlag, "", "Leanix configuration name created on the workspace")
	flag.Parse()
	// Let flags overwrite configs in viper
	err := viper.BindPFlags(flag.CommandLine)
	if err != nil {
		logger.Error("Error overwriting configs in viper", err)
		return err
	}
	// Check for config values in env vars
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	if viper.GetString(utils.LxWorkspaceFlag) == "" {
		return fmt.Errorf("%s flag must be set", utils.LxWorkspaceFlag)
	}
	if viper.GetBool(utils.EnableCustomStorageFlag) {
		if viper.GetString(utils.StorageBackendFlag) == "none" || viper.GetString(utils.StorageBackendFlag) == "" {
			return fmt.Errorf("%s flag must be set since %s is enabled", utils.StorageBackendFlag, utils.EnableCustomStorageFlag)
		}

		if viper.GetString(utils.StorageBackendFlag) == "azureblob" {
			if viper.GetString(utils.AzureAccountNameFlag) == "" {
				return fmt.Errorf("%s flag must be set", utils.AzureAccountNameFlag)
			}
			if viper.GetString(utils.AzureAccountKeyFlag) == "" {
				return fmt.Errorf("%s flag must be set", utils.AzureAccountKeyFlag)
			}
			if viper.GetString(utils.AzureContainerFlag) == "" {
				return fmt.Errorf("%s flag must be set", utils.AzureContainerFlag)
			}
		}
	}
	if viper.GetBool(utils.IrisFlag) {
		if viper.GetString(utils.ConfigurationNameFlag) == "" {
			return fmt.Errorf("%s flag must be set", utils.ConfigurationNameFlag)
		}
		return nil
	}
	if viper.GetString(utils.IntegrationAPIDatasourceNameFlag) == "" {
		return fmt.Errorf("%s flag must be set", utils.IntegrationAPIDatasourceNameFlag)
	}
	return nil
}
