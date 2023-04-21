package iris

import (
	models "github.com/leanix/leanix-k8s-connector/pkg/iris/models/namespace"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models/workload"
)

type ConfigService interface {
	GetConfiguration(configurationName string) ([]byte, error)
	GetNamespaceScanResults(configurationId string) ([]models.DiscoveryEvent, error)
	GetWorkloadScanResults(configurationId string) ([]workload.DiscoveryEvent, error)
}

type configService struct {
	irisApi API
}

func NewConfigService(irisApi API) ConfigService {
	return &configService{
		irisApi: irisApi,
	}
}

func (a *configService) GetConfiguration(configurationName string) ([]byte, error) {
	return a.irisApi.GetConfiguration(configurationName)
}

func (a *configService) GetNamespaceScanResults(configurationId string) ([]models.DiscoveryEvent, error) {
	return a.irisApi.GetNamespaceScanResults(configurationId)
}

func (a *configService) GetWorkloadScanResults(configurationId string) ([]workload.DiscoveryEvent, error) {
	return a.irisApi.GetWorkloadScanResults(configurationId)
}
