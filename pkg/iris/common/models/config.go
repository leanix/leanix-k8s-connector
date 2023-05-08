package models

type KubernetesConfig struct {
	ID                    string   `json:"id"`
	Cluster               string   `json:"cluster"`
	BlackListedNamespaces []string `json:"blacklistedNamespaces"`
	DiscoveryMode         string   `json:"discoveryMode"`
}
