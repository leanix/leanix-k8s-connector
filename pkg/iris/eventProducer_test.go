package iris

import (
	"github.com/golang/mock/gomock"
	"github.com/leanix/leanix-k8s-connector/pkg/iris/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_eventProducer_filter_created(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockApi := NewMockAPI(ctrl)
	newData := map[string]models.Data{

		"testId1": {
			models.ClusterEcst{
				Namespace:   "testNamespace1",
				Deployments: make([]models.DeploymentEcst, 0),
				Name:        "testCluster1",
				Os:          "testOs1",
				K8sVersion:  "1.0.0",
				NoOfNodes:   "1",
			}},
	}
	oldData := map[string]models.DiscoveryEvent{
		"testId2": {
			Body: models.DiscoveryBody{
				State: models.State{
					Name:   "",
					Source: "",
					Time:   "",
					Data:   models.Data{},
				},
			},
		},
	}
	//oldData map[string]models.DiscoveryEvent
	p := &eventProducer{
		irisApi:     mockApi,
		runId:       "testRunId",
		configId:    "testConfigId",
		workspaceId: "testWorkspaceId",
	}
	created, updated, _, err := p.filterForChangedItems(newData, oldData)

	assert.NoError(t, err)
	assert.Empty(t, updated)
	assert.Len(t, created, 1)
	assert.Equal(t, "testNamespace1", created[0].Body.State.Name)
}

func Test_eventProducer_filter_updated_no_change(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockApi := NewMockAPI(ctrl)
	newData := map[string]models.Data{

		"testId1": {
			models.ClusterEcst{
				Namespace:   "testNamespace1",
				Deployments: make([]models.DeploymentEcst, 0),
				Name:        "testCluster1",
				Os:          "testOs1",
				K8sVersion:  "1.0.0",
				NoOfNodes:   "1",
			}},
	}
	oldData := map[string]models.DiscoveryEvent{
		"testId1": {
			Body: models.DiscoveryBody{
				State: models.State{
					Name:   "",
					Source: "",
					Time:   "",
					Data: models.Data{
						Cluster: models.ClusterEcst{
							Namespace:   "testNamespace1",
							Deployments: make([]models.DeploymentEcst, 0),
							Name:        "testCluster1",
							Os:          "testOs1",
							K8sVersion:  "1.0.0",
							NoOfNodes:   "1",
						},
					},
				},
			},
		},
	}
	//oldData map[string]models.DiscoveryEvent
	p := &eventProducer{
		irisApi:     mockApi,
		runId:       "testRunId",
		configId:    "testConfigId",
		workspaceId: "testWorkspaceId",
	}
	created, updated, filteredData, err := p.filterForChangedItems(newData, oldData)

	assert.NoError(t, err)
	assert.Empty(t, updated)
	assert.Empty(t, created)
	assert.Empty(t, filteredData)
}

func Test_eventProducer_filter_updated_changed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockApi := NewMockAPI(ctrl)
	newData := map[string]models.Data{

		"testId1": {
			models.ClusterEcst{
				Namespace:   "testNamespace1",
				Deployments: make([]models.DeploymentEcst, 0),
				Name:        "testCluster1",
				Os:          "testOs1",
				K8sVersion:  "1.5.0",
				NoOfNodes:   "1",
			}},
	}
	oldData := map[string]models.DiscoveryEvent{
		"testId1": {
			Body: models.DiscoveryBody{
				State: models.State{
					Name:   "",
					Source: "",
					Time:   "",
					Data: models.Data{
						Cluster: models.ClusterEcst{
							Namespace:   "testNamespace1",
							Deployments: make([]models.DeploymentEcst, 0),
							Name:        "testCluster1",
							Os:          "testOs1",
							K8sVersion:  "1.0.0",
							NoOfNodes:   "1",
						},
					},
				},
			},
		},
	}
	//oldData map[string]models.DiscoveryEvent
	p := &eventProducer{
		irisApi:     mockApi,
		runId:       "testRunId",
		configId:    "testConfigId",
		workspaceId: "testWorkspaceId",
	}
	created, updated, filteredData, err := p.filterForChangedItems(newData, oldData)

	assert.NoError(t, err)
	assert.Empty(t, created)
	assert.Len(t, updated, 1)
	assert.Equal(t, "1.5.0", updated[0].Body.State.Data.Cluster.K8sVersion)
	assert.Empty(t, filteredData)
}
