package iris

import (
	"github.com/leanix/leanix-k8s-connector/pkg/logger"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setup() {
	logger.Init()
}

func TestGetConfiguration200(t *testing.T) {
	setup()
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.Method, "GET")
		assert.Contains(t, req.URL.String(), "/services/vsm-iris/v1/configurations/kubernetes/test-config")
		// Send response to be tested
		rw.Write([]byte(`OK`))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Use Client & URL from our local test server
	api := NewApi(server.Client(), "kind-test", server.URL)

	configuration, err := api.GetConfiguration("test-config", "test-token")
	assert.NoError(t, err)
	assert.Equal(t, "OK", string(configuration))
}

func TestPostResults200(t *testing.T) {
	setup()
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.Method, "POST")
		assert.Contains(t, req.URL.String(), "/services/vsm-iris/v1/results")
		requestData, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, "test-results", string(requestData))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Use Client & URL from our local test server
	api := NewApi(server.Client(), "kind-test", server.URL)
	results := []byte("test-results")
	err := api.PostEcstResults(results, "test-token")
	assert.NoError(t, err)
}

func TestPostResultsError(t *testing.T) {
	setup()
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(`Exception`))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Use Client & URL from our local test server
	api := NewApi(server.Client(), "kind-test", server.URL)
	results := []byte("test-results")
	err := api.PostEcstResults(results, "test-token")
	assert.Equal(t, "posting ECST results status [500 Internal Server Error] could not be processed: 'Exception'", err.Error())
}
