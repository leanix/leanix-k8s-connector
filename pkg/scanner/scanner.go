package scanner

import (
	"bytes"
	"context"
	"reflect"
	"strings"

	"github.com/leanix/leanix-k8s-connector/pkg/kubernetes"
	"github.com/leanix/leanix-k8s-connector/pkg/leanix"
	"github.com/leanix/leanix-k8s-connector/pkg/mapper"
	"github.com/op/go-logging"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var log = logging.MustGetLogger("leanix-k8s-connector")

func ScanKubernetesCurrent(startResponse *leanix.SelfStartResponse, blacklistedNamespacesList []string, config *rest.Config, logger *bytes.Buffer) ([]mapper.KubernetesObject, error) {

	kubernetesAPI, err := kubernetes.NewAPI(config)
	if err != nil {
		return nil, err
	}
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Error(err)
	}

	log.Debug("Get blacklist namespaces list...")
	blacklistedNamespaces, err := kubernetesAPI.BlackListNamespaces(blacklistedNamespacesList)
	if err != nil {
		return nil, err
	}
	log.Debug("Getting blacklist namespaces list done.")
	log.Infof("Namespace blacklist: %v", reflect.ValueOf(blacklistedNamespaces).MapKeys())

	resourcesList, err := ServerPreferredListableResources(kubernetesAPI.Client.Discovery())
	if err != nil {
		return nil, err
	}
	groupVersionResources, err := discovery.GroupVersionResources(resourcesList)
	if err != nil {
		return nil, err
	}
	_, err = leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Discovery of Version Resources is done. Moving on to mapping nodes")
	if err != nil {
		log.Errorf("KubernetesScan: Before scan: Failed to update progress[%s] to Integration Hub", leanix.IN_PROGRESS, err)
	}
	log.Debug("Listing nodes...")
	nodes, err := kubernetesAPI.Nodes()
	if err != nil {
		return nil, err
	}
	log.Debug("Listing nodes done.")
	log.Debug("Map nodes to Kubernetes object")
	clusterKubernetesObject, err := mapper.MapNodes(
		startResponse.ConnectorConfiguration.ClusterName,
		nodes,
	)
	if err != nil {
		return nil, err
	}
	kubernetesObjects := make([]mapper.KubernetesObject, 0)
	kubernetesObjects = append(kubernetesObjects, *clusterKubernetesObject)
	_, statusErr := leanix.UpdateInProgressStatus(startResponse.ProgressCallbackUrl, "Mapping nodes is done. Moving on to collecting kubernetes objects from Version Resources.")
	if statusErr != nil {
		log.Errorf("KubernetesScan: After Scan: Failed to update progress[%s] to Integration Hub", leanix.IN_PROGRESS, statusErr)

	}
	resourceGroupWhitelist := map[string]map[string]interface{}{
		"": map[string]interface{}{
			"serviceaccounts":        struct{}{},
			"services":               struct{}{},
			"nodes":                  struct{}{},
			"pods":                   struct{}{},
			"namespaces":             struct{}{},
			"configmaps":             struct{}{},
			"persistentvolumes":      struct{}{},
			"persistentvolumeclaims": struct{}{},
			"replicationcontrollers": struct{}{},
		},
		"apps": map[string]interface{}{
			"deployments":  struct{}{},
			"statefulsets": struct{}{},
			"daemonsets":   struct{}{},
			"replicasets":  struct{}{},
		},
		"apiextensions.k8s.io": map[string]interface{}{
			"customresourcedefinitions": struct{}{},
		},
		"rbac.authorization.k8s.io": map[string]interface{}{
			"clusterrolebindings": struct{}{},
			"rolebindings":        struct{}{},
			"clusterroles":        struct{}{},
			"roles":               struct{}{},
		},
		"networking.k8s.io": map[string]interface{}{
			"ingresses":       struct{}{},
			"networkpolicies": struct{}{},
		},
		"autoscaling": map[string]interface{}{
			"horizontalpodautoscalers": struct{}{},
		},
		"policy": map[string]interface{}{
			"podsecuritypolicies": struct{}{},
		},
		"storage.k8s.io": map[string]interface{}{
			"storageclasses": struct{}{},
		},
		"batch": map[string]interface{}{
			"cronjobs": struct{}{},
			"jobs":     struct{}{},
		},
	}

	for gvr := range groupVersionResources {
		if _, ok := resourceGroupWhitelist[gvr.Group][gvr.Resource]; !ok {
			log.Debugf("Not scanning resource %s", strings.Join([]string{gvr.Group, gvr.Version, gvr.Resource}, "/"))
			continue
		}
		instances, err := dynClient.Resource(gvr).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, i := range instances.Items {
			if _, ok := blacklistedNamespaces[i.GetNamespace()]; ok {
				continue
			}
			nko := mapper.KubernetesObject{
				Type: i.GetKind(),
				ID:   string(i.GetUID()),
				Data: i.Object,
			}
			kubernetesObjects = append(kubernetesObjects, nko)
		}
	}
	return kubernetesObjects, nil
}

func ServerPreferredListableResources(d discovery.DiscoveryInterface) ([]*metav1.APIResourceList, error) {
	all, err := discovery.ServerPreferredResources(d)
	return discovery.FilteredBy(discovery.ResourcePredicateFunc(func(groupVersion string, r *metav1.APIResource) bool {
		return strings.Contains(r.Verbs.String(), "list")
	}), all), err
}
