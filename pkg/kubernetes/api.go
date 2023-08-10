package kubernetes

import (
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// API is an optionated facade for the Kubernetes api
type API struct {
	Client kubernetes.Interface
}

// NewAPI creates a new Kubernetes api client
func NewAPI(config *rest.Config) (*API, error) {
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &API{
		Client: clientset,
	}, nil
}

type GetKubernetesAPI func(config *rest.Config) (*API, error)

// NamespaceBlacklistFieldSelector builds a Field Selector string to filter the response to not
// include namespaces that belong to the blacklisted namespaces
func NamespaceBlacklistFieldSelector(blacklistedNamespaces []string) string {
	namespaceSelectors := Prefix(blacklistedNamespaces, "metadata.name!=")
	return strings.Join(namespaceSelectors, ",")
}

// Prefix returns a new list where all items are prefixed with the string given as prefix
func Prefix(l []string, p string) []string {
	r := make([]string, 0)
	for _, e := range l {
		r = append(r, (p + e))
	}
	return r
}
