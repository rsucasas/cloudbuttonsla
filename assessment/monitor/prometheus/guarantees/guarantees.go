/*
Copyright 2018 Atos

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package guarantees contains a the monitorization of specific guarantees.
package guarantees

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/common/log"
)

type QueryResp struct {
	Status string `json:"status"`
	Data   data   `json:"data"`
}

type data struct {
	ResultType string   `json:"resultType"`
	Results    []result `json:"result"`
}

type result struct {
	Metric metric        `json:"metric"`
	Values []interface{} `json:"value"`
}

type metric struct {
	Name             string `json:"__name__"`
	Instance         string `json:"instance"`
	Job              string `json:"job"`
	ExportedInstance string `json:"exported_instance"`
	ExecutorId       string `json:"executor_id"`
	ExportedJob      string `json:"job_id"`
	Function         string `json:"function_name"`
	Call             string `json:"call_id"`
}

/*
CheckGuarantee checks specifi guarantees
*/
func CheckGuarantee(gname string, prometheusRootURL string, execID string, resource string) []string {
	switch strings.ToLower(gname) {
	case "notstarted":
		return notStarted(prometheusRootURL, execID, resource)
	default:
		return []string{}
	}
}

/**
 * Query single query to prometheus
 */
func Query(m string, prometheusRootURL string) (QueryResp, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusRootURL, m)
	return request(url)
}

// http request
func request(url string) (QueryResp, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return QueryResp{}, err
	}
	defer resp.Body.Close()

	var result QueryResp
	if resp.Status[0] != '2' { // StatusCode < 200 || resp.StatusCode >= 300
		log.Errorf("%s GET %s", resp.Status, url)
	}
	log.Infof("%d %s", resp.StatusCode, url)

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Errorf("Error decoding prometheus output: %s", err.Error())
		return QueryResp{}, err
	}

	log.Infof("http request response: %v", result)
	return result, nil
}
