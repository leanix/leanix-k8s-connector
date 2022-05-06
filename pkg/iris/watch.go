package iris

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-test/deep"
	"github.com/op/go-logging"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var log = logging.MustGetLogger("leanix-k8s-connector")

func WatchKubernetes(config *rest.Config, workspaceId string, accessToken string) (string, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}
	factory := informers.NewSharedInformerFactory(clientset, 0)
	informer := factory.Apps().V1().Deployments().Informer()
	stopper := make(chan struct{})
	defer close(stopper)
	defer runtime.HandleCrash()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})
	go informer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return "", fmt.Errorf("Timed out waiting for caches to sync")
	}
	<-stopper
	return "test", nil
}

// onAdd is the function executed when the kubernetes informer notified the
// presence of a new kubernetes node in the cluster
func onAdd(obj interface{}) {
	// Cast the obj as node
	deployment := obj.(*appsv1.Deployment)
	mappedDiscoveryItem := deployment
	//convert node to bytes
	deploymentBytes, err := json.Marshal(mappedDiscoveryItem)
	if err != nil {
		log.Panic(err.Error())
	}
	response, err := http.Post("http://127.0.0.1:8080/k8sConnector/postResults", "application/json", bytes.NewBuffer(deploymentBytes))
	if err != nil {
		log.Info(err)
	}
	log.Info(response)

}
func onUpdate(oldObj interface{}, newObj interface{}) {

	if !equality.Semantic.DeepEqual(oldObj, newObj) {
		diff := deep.Equal(oldObj, newObj)
		// Cast the obj as node
		//deployment := newObj.(*appsv1.Deployment)
		//mappedDiscoveryItem, err := MapSoftwareArtifact(clusterName, workspaceId, deployment)

		//convert node to bytes
		deploymentBytes, err := json.Marshal(diff)
		if err != nil {
			log.Panic(err.Error())
		}
		response, err := http.Post("http://127.0.0.1:8080/k8sConnector/postResults", "application/json", bytes.NewBuffer(deploymentBytes))
		if err != nil {
			log.Info(err)
		}
		log.Info(response)
	}

}
func onDelete(obj interface{}) {
	// Cast the obj as node
	deployment := obj.(*appsv1.Deployment)
	//deletedDeployment, err := MapDeleteSoftwareArtifact(clusterName, workspaceId, deployment)

	//convert node to bytes
	deploymentBytes, err := json.Marshal(deployment)
	if err != nil {
		log.Panic(err.Error())
	}
	response, err := http.Post("http://127.0.0.1:8080/k8sConnector/postResults", "application/json", bytes.NewBuffer(deploymentBytes))
	if err != nil {
		log.Info(err)
	}
	log.Info(response)
}
