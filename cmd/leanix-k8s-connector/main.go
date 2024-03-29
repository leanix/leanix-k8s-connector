package main

import (
	"fmt"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/common/models"
	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	"path/filepath"
	"strings"

	"github.com/leanix/leanix-k8s-connector/pkg/iris"
	"github.com/leanix/leanix-k8s-connector/pkg/leanix"
	"github.com/leanix/leanix-k8s-connector/pkg/utils"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	logger.Init()
	err := parseFlags()
	if err != nil {
		logger.Error("Error occurred when parsing flags.", err)
	}

	var config *restclient.Config
	if viper.GetBool(utils.LocalFlag) {
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		if err != nil {
			logger.Errorf("Failed to load kube config. Running locally?\n%s", err)
		}
	} else {
		config, err = restclient.InClusterConfig()
		if err != nil {
			logger.Errorf("Failed to load kube config. Running in Kubernetes?\n%s", err)
		}
	}
	apiHostFqdn := viper.GetString(utils.IntegrationAPIFqdnFlag)
	apiToken := viper.GetString(utils.IntegrationAPITokenFlag)
	if apiHostFqdn == "" {
		apiHostFqdn = viper.GetString(utils.ApiHostFlag)
	}
	if apiToken == "" {
		apiToken = viper.GetString(utils.ApiTokenFlag)
	}
	accessToken, err := leanix.Authenticate(apiHostFqdn, apiToken)
	if err != nil {
		logger.Error("Error occurred when authenticating.", err)
		logger.Info("Failed to authenticate. Terminating..")
		return
	}
	if viper.GetBool(utils.IrisFlag) {
		logger.Info("Enabled Iris")
		runId := models.GenerateRunId()
		irisScanner := iris.NewScanner(
			"Iris Integration",
			apiHostFqdn,
			runId,
			accessToken,
			viper.GetString(utils.LxWorkspaceFlag),
		)

		err = irisScanner.Scan(kubernetes.NewAPI, config, viper.GetString(utils.ConfigurationNameFlag))
		if err != nil {
			logger.Error("Failed to scan Kubernetes via vsm-iris.", err)
		}
	} else {
		logger.Error("Using deprecated configuration. Please set the iris flag to true.", err)
	}
}

func parseFlags() error {
	flag.Bool(utils.EnableCustomStorageFlag, false, "Disable/enable custom storage backend option")
	flag.String(utils.AzureAccountNameFlag, "", "Azure storage account name")
	flag.String(utils.AzureAccountKeyFlag, "", "Azure storage account key")
	flag.String(utils.AzureContainerFlag, "", "Azure storage account container")
	flag.String(utils.LocalFilePathFlag, ".", "path to place the ldif file when using local file storage backend")
	flag.Bool(utils.VerboseFlag, false, "verbose log output")
	flag.String(utils.IntegrationAPIDatasourceNameFlag, "", "LeanIX Integration Hub Datasource name created on the workspace")
	flag.String(utils.IntegrationAPIFqdnFlag, "", "LeanIX Instance FQDN - deprecated flag")
	flag.String(utils.IntegrationAPITokenFlag, "", "LeanIX API token - deprecated flag")
	flag.String(utils.ApiHostFlag, "app.leanix.net", "LeanIX Instance FQDN")
	flag.String(utils.ApiTokenFlag, "", "LeanIX API token")
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
