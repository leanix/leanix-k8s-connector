package kubernetes

import (
	"context"
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var replacer = strings.NewReplacer(
	"*", "\\\\*",
)

// Namespaces gets the list of blacklisted namespaces
func (k *API) Namespaces(blacklistedNamespaces []string) (*v1.NamespaceList, error) {
	namespaces, err := k.Client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return namespaces, err
}

// Namespaces gets the list of blacklisted namespaces
func (k *API) BlackListNamespaces(blacklistedNamespaces []string) (map[string]interface{}, error) {
	namespaces, err := k.Namespaces(blacklistedNamespaces)
	if err != nil {
		return nil, err
	}
	blacklist := make(map[string]interface{}, 0)
	for _, m := range blacklistedNamespaces {
		for _, n := range namespaces.Items {
			t := replacer.Replace(strings.Trim(m, ","))
			re := regexp.MustCompile(t)
			if re.MatchString(n.Name) {
				blacklist[n.Name] = struct{}{}
			}
		}
	}

	return blacklist, nil
}
