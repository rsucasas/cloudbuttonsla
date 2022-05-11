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

// Package rabbitpushgnotifier contains a simple ViolationsNotifier that send violations to a rabbit queue and also
// the queries to Prometheus needed to get all the information about the violation.
package rabbitpushgnotifier

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
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
	Name string `json:"__name__"`
	//Instance         string `json:"instance"`
	ExportedInstance string `json:"exported_instance"`
	Job              string `json:"job_id"`
	Function         string `json:"function_name"`
	Call             string `json:"call_id"`
}

///////////////////////////////////////////////////////////////////////////////

/**
 * Query single query to prometheus
 */
func Query(m string, prometheusRootURL string) (QueryResp, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusRootURL, m)
	log.Debugf("guarantees [Query] url: %s", url)
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
	log.Debugf("%d %s", resp.StatusCode, url)

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Errorf("Error decoding prometheus output: %s", err.Error())
		return QueryResp{}, err
	}

	log.Debugf("http request response: %v", result)
	return result, nil
}

/**
 * ChecknotStartedFunctions queries Prometheus for the functions that didn't start from a job.
 * Returns a []string with the instances ids of no started functions, e.g. []string{"935882-0-A000-00002", "935882-0-A000-00003"}
 */
func ChecknotStartedFunctions(prometheusRootURL string, jobId string, resource string) []string {
	return notStarted(prometheusRootURL, jobId, resource)
}

/**
 * notStarted this function queries Prometheus for the functions that didn't start from a job.
 * Returns a []string with the instances ids of no started functions, e.g. []string{"935882-0-A000-00002", "935882-0-A000-00003"}
 */
func notStarted(prometheusRootURL string, jobId string, resource string) []string {
	log.Debugf("guarantees [notStarted] jobId:  %s , resource: %s, prometheus URL: %s", jobId, resource, prometheusRootURL)

	// 1. get total calls: "job_total_calls{job_id='935882-1-A000'}"
	queryStr := "job_total_calls{job_id='" + jobId + "'}"

	log.Debugf("guarantees [notStarted] Getting 'jobTotalCalls' value from Prometheus > queryStr: %s", queryStr)
	resQ, err := Query(queryStr, prometheusRootURL)
	if err == nil && len(resQ.Data.Results) > 0 && len(resQ.Data.Results[0].Values) > 1 {

		jobTotalCalls, err := strconv.Atoi(resQ.Data.Results[0].Values[1].(string))
		if err == nil {
			job_id := resQ.Data.Results[0].Metric.Job

			log.Infof("guarantees [notStarted] > 'jobTotalCalls' value for [function_name="+resource+" ,call_id="+jobId+"]: %s", resQ.Data.Results[0].Values[1].(string))

			list := []string{}

			// 2. get all 'function_start'
			// generate a value for the 'call_id' values, e.g. "935882-0-A000-00002" ==> "935882-0-A000-.*"
			callId := getCallIdValuesRegex(jobId, job_id)

			// create query string. Example: function_start{function_name="test_function_1",exported_instance=~"935882-0-.*"}
			queryStr = "function_start{call_id=~'" + callId + "'}"
			log.Debugf("guarantees [notStarted] > Getting all 'function_start' values from Prometheus > queryStr: %s", queryStr)

			// execut query
			resQ, err = Query(queryStr, prometheusRootURL)
			if err == nil && len(resQ.Data.Results) > 0 {
				list = processQuery(resQ, jobTotalCalls, jobId, job_id) // return []string{}
			} else if err == nil && len(resQ.Data.Results) == 0 {
				log.Warnf("guarantees [notStarted] > query with no results = 'function_start' not found for defined callId values")
			} else if err != nil {
				log.Errorf("guarantees [notStarted] > Error in call to Query [function_start] function: %s", err.Error())
			}

			return list
		} else {
			log.Errorf("guarantees [notStarted] > Error converting response value to jobTotalCalls: %s", err.Error())
		}
	} else if err == nil && len(resQ.Data.Results) == 0 {
		log.Warnf("guarantees [notStarted] > No results obtained from call to Query [job_total_calls] function: [len(resQ.Data.Results) == 0]")
	} else {
		log.Errorf("guarantees [notStarted] > Error in call to Query [job_total_calls] function: %s", err.Error())
	}

	return []string{}
}

// generates a call_id
func getCallIdValuesRegex(jobId string, job_id string) string {
	//return jobId + "-" + job_id + "-.*"
	return job_id + "-.*"
}

// processQuery process the results and returns a list of not started callIds
func processQuery(q QueryResp, t int, jobId string, job_id string) []string {
	list := []string{}

	for i := 0; i < t; i++ {
		callIdValue := getCallIdValue2(jobId, job_id, generateCallId2(i))
		found := false

		for _, r := range q.Data.Results {
			if r.Metric.Call == callIdValue {
				found = true
				break
			}
		}

		if !found {
			list = append(list, callIdValue)
		}
	}

	return list
}

// generates a call_id
func generateCallId2(i int) string {
	if i < 10 {
		return "0000" + strconv.Itoa(i)
	} else if i < 100 {
		return "000" + strconv.Itoa(i)
	} else if i < 1000 {
		return "00" + strconv.Itoa(i)
	} else if i < 10000 {
		return "0" + strconv.Itoa(i)
	}

	return strconv.Itoa(i)
}

// getCallIdValue
func getCallIdValue2(jobId string, job_id string, callId string) string {
	//return jobId + "-" + job_id + "-" + callId
	return job_id + "-" + callId
}
