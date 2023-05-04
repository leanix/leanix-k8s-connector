package services

import (
	"github.com/leanix/leanix-k8s-connector/pkg/iris/common/models"
)

type ConfigService interface {
	GetConfiguration(configurationName string) ([]byte, error)
	GetScanResults(configurationId string) ([]models.DiscoveryEvent, error)
}

type configService struct {
	irisApi IrisApi
}

func NewConfigService(irisApi IrisApi) ConfigService {
	return &configService{
		irisApi: irisApi,
	}
}

func (a *configService) GetConfiguration(configurationName string) ([]byte, error) {
	return a.irisApi.GetConfiguration(configurationName)
}

func (a *configService) GetScanResults(configurationId string) ([]models.DiscoveryEvent, error) {
	return a.irisApi.GetScanResults(configurationId)
}
