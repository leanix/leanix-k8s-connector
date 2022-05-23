package leanix

import "k8s.io/client-go/rest"

type ClientInterface interface {
	GetConfiguration(configurationName string, accessToken string) ([]byte, error)
	PostResults(results []byte, accessToken string) string
	ScanKubernetes(config *rest.Config, workspaceId string, configurationName string, accessToken string) error
}

type Integration struct {
	Kind   string
	Uri    string
	Client ClientInterface
}
