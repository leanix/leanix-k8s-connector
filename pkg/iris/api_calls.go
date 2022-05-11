package iris

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func PostResults(results []byte, accessToken string) (string, error) {
	resultUrl := "http://localhost:8080/results"
	postReq, err := http.NewRequest("POST", resultUrl, nil)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+accessToken)
	postReq.Body = ioutil.NopCloser(bytes.NewBuffer(results))
	if err != nil {
		log.Infof("Post results request: Error while posting results: %s", resultUrl)
		return postReq.URL.String(), err
	}
	resp, err := http.DefaultClient.Do(postReq)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		responseData, readErr := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if readErr != nil {
			return "", readErr
		}
		err := fmt.Errorf("Posting results status[%s]could not be processed: '%s'\n", resp.Status, responseData)
		return "", err
	}
	return "", nil
}
