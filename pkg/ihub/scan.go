package ihub

import (
	"github.com/leanix/leanix-k8s-connector/pkg/leanix"
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	"github.com/leanix/leanix-k8s-connector/pkg/mapper"
	"github.com/leanix/leanix-k8s-connector/pkg/newmapper"
	"github.com/leanix/leanix-k8s-connector/pkg/scanner"
	"github.com/leanix/leanix-k8s-connector/pkg/storage"
	"github.com/leanix/leanix-k8s-connector/pkg/utils"
	"github.com/leanix/leanix-k8s-connector/pkg/version"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
)

func KubernetesScan(config *rest.Config, viper *viper.Viper) (response *leanix.SelfStartResponse, err error) {
	logger.Info("----------Attempting to Self Start via Integration Hub----------")

	accessToken, err := leanix.Authenticate(viper.GetString(utils.IntegrationAPIFqdnFlag), viper.GetString(utils.IntegrationAPITokenFlag))
	if err != nil {
		logger.Error("Error when authenticating.", err)
		logger.Info("Failed to authenticate. Terminating..")
		return nil, err
	}
	logger.Info("Integration Hub authentication successful.")
	startResponse, err := leanix.SelfStartRun(viper.GetString(utils.IntegrationAPIFqdnFlag), accessToken, viper.GetString(utils.IntegrationAPIDatasourceNameFlag))
	if err != nil {
		logger.Info("Main: Failed to start Integration Hub. Terminating..")
		return startResponse, err
	}

	if startResponse != nil {
		logger.Infof("Successfully self started via Integration Hub. Progress call back - %s", startResponse.ProgressCallbackUrl)
		_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Successfully self started via Integration Hub. Connector is in progress")
		if err != nil {
			logger.Errorf("KubernetesScan: Failed to update progress[%s] to Integration Hub: '%s'", leanix.IN_PROGRESS, err)
		}
	}

	logger.Info("Getting connector config...")
	logger.Info("----------Start----------")
	logger.Infof("LeanIX Kubernetes connector version: %s", version.VERSION)
	logger.Infof("LeanIX Kubernetes connector build version: %s", version.FULL_VERSION)
	logger.Infof("LeanIX integration version: %s", startResponse.BindingKey.LXVersion)
	logger.Infof("LeanIX connector id: %s", startResponse.BindingKey.ConnectorId)
	logger.Infof("LeanIX connector type: %s", startResponse.BindingKey.ConnectorType)
	logger.Infof("LeanIX connector processing direction: %s", startResponse.BindingKey.ProcessingDirection)
	logger.Infof("Target LeanIX workspace: %s", viper.GetString(utils.LxWorkspaceFlag))
	logger.Infof("Target Kubernetes cluster name: %s", startResponse.ConnectorConfiguration.ClusterName)

	logger.Debugf("Kubernetes master from config: %s", config.Host)
	kubernetesObjects := make([]mapper.KubernetesObject, 0)
	if startResponse.BindingKey.ConnectorId == "leanix-k8s-v3-connector" {
		logger.Info("Using new k8s-v3 mapper")
		mappedObject, err := newmapper.ScanKubernetes(startResponse.ConnectorConfiguration.ClusterName, config)
		if err != nil {
			return nil, err
		}
		kubernetesObjects = append(kubernetesObjects, mappedObject...)

		_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Mapping of kubernetes objects with new version done. Preparing ldif.")
		if err != nil {
			logger.Errorf("New Kubernetes Scan: After Mapping: Failed to update progress[%s] to Integration Hub"+leanix.IN_PROGRESS, err)
		}
	} else {
		logger.Info("Using current k8s mapper")
		blacklistedNamespacesList := viper.GetStringSlice(utils.BlacklistNamespacesFlag)
		mappedObject, err := scanner.ScanKubernetesCurrent(startResponse, blacklistedNamespacesList, config)
		if err != nil {
			return startResponse, err
		}
		_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Mapping of kubernetes objects with old version done. Preparing ldif.")
		if err != nil {
			logger.Errorf("Current Kubernetes Scan: After Mapping: Failed to update progress[%s] to Integration Hub"+leanix.IN_PROGRESS, err)
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
		LxWorkspace:         viper.GetString(utils.LxWorkspaceFlag),
		Description:         "Map Kubernetes objects to LeanIX Fact Sheets",
		CustomFields:        customFields,
		Content:             kubernetesObjects,
	}

	_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Successfully collected required kubernetes data.")
	if err != nil {
		logger.Errorf("KubernetesScan: After mapping: Failed to progress[%s] to Integration Hub"+leanix.IN_PROGRESS, err)
	}
	logger.Debug("Marshal ldif")
	ldifByte, err := storage.Marshal(ldif)
	if err != nil {
		return nil, err
	}

	if err != nil {
		_, statusErr := leanix.UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "Failed to marshal ldif")
		if statusErr != nil {
			logger.Errorf("KubernetesScan: Marshal ldif error: Failed to progress[%s] to Integration Hub"+leanix.FAILED, statusErr)
		}
		return startResponse, err
	}

	if viper.GetBool(utils.EnableCustomStorageFlag) {
		_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Uploading ldif to configured storage backend - "+viper.GetString("storage-backend"))
		if err != nil {
			logger.Infof("Failed to progress[%s] to Integration Hub", leanix.IN_PROGRESS)
		}

		logger.Infof("Upload %s to %s", storage.LdifFileName, viper.GetString("storage-backend"))
		azureOpts := storage.AzureBlobOpts{
			AccountName: viper.GetString(utils.AzureAccountNameFlag),
			AccountKey:  viper.GetString(utils.AzureAccountKeyFlag),
			Container:   viper.GetString(utils.AzureContainerFlag),
		}
		localFileOpts := storage.LocalFileOpts{
			Path: viper.GetString(utils.LocalFilePathFlag),
		}
		uploader, err := storage.NewBackend(viper.GetString("storage-backend"), &azureOpts, &localFileOpts)
		if err != nil {
			_, statusErr := leanix.UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "Failed to create uploader for backend storage")
			if statusErr != nil {
				logger.Errorf("KubernetesScan: Failed to create uploader for backend storage: Failed to progress[%s] to Integration Hub"+leanix.FAILED, statusErr)
			}
			return nil, err
		}
		err = uploader.UploadLdif(ldifByte, storage.LdifFileName+viper.GetString(utils.LxWorkspaceFlag)+storage.LdifFileExtension)
		if err != nil {
			_, statusErr := leanix.UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "Failed to upload ldif to backend storage configured storage backend - "+viper.GetString("storage-backend"))
			if statusErr != nil {
				logger.Errorf("KubernetesScan: Upload LDIF error: Failed to progress[%s] to Integration Hub"+leanix.FAILED, statusErr)
			}
			return nil, err
		}

		_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Successfully uploaded ldif to configured storage backend - "+viper.GetString("storage-backend"))
		if err != nil {
			logger.Errorf("Failed to update progress[%s] to Integration Hub"+leanix.IN_PROGRESS, err)
		}
		logger.Info("-----------Done Uploading to custom storage backend-----------")
	} else {
		logger.Infof("Skipping uploading LDIF and log file to custom storage backend. reason - flag is disabled")
	}

	_, err = leanix.UploadLdif(startResponse.LdifResultUrl, ldifByte)
	if err != nil {
		logger.Debug("Failed to upload ldif to Integration Hub ldif SAS Url")
		_, err := leanix.UpdateFailedProgressStatus(startResponse.ProgressCallbackUrl, "Failed to upload ldif to Integration Hub ldif SAS Url")
		logger.Error("Error when uploading log file to iHub", err)
		return nil, err
	}
	_, err = leanix.UpdateProgress(startResponse.ProgressCallbackUrl, leanix.FINISHED, "")
	if err != nil {
		logger.Errorf("Failed to update progress[%s] to Integration Hub"+leanix.FINISHED, err)
	}
	logger.Debug("-----------End-----------")
	logger.Info("-----------End-----------")
	return startResponse, err
}
