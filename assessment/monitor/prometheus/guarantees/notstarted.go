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
	"strconv"

	"github.com/prometheus/common/log"
)

///////////////////////////////////////////////////////////////////////////////

/**
 * notStarted this function queries Prometheus for the functions that didn't start from a job.
 * Returns a []string with the instances ids of no started functions, e.g. []string{"935882-0-A000-00002", "935882-0-A000-00003"}
 */
func notStarted(prometheusRootURL string, execID string, resource string) []string {
	log.Infof("guarantees [notStarted] execID:  %s , resource: %s, prometheus URL: %s", execID, resource, prometheusRootURL)

	// 1. get total calls: "job_total_calls{function_name='test_function_1',executor_id='935882-0'}""
	queryStr := "job_total_calls{function_name='" + resource + "',executor_id='" + execID + "'}"

	log.Infof("guarantees [notStarted] Getting 'jobTotalCalls' value from Prometheus > queryStr: %s", queryStr)
	resQ, err := Query(queryStr, prometheusRootURL)
	if err == nil && len(resQ.Data.Results) > 0 && len(resQ.Data.Results[0].Values) > 1 {

		jobTotalCalls, err := strconv.Atoi(resQ.Data.Results[0].Values[1].(string))
		if err == nil {
			job_id := resQ.Data.Results[0].Metric.ExportedJob

			log.Infof("guarantees [notStarted] > 'jobTotalCalls' value: %s", resQ.Data.Results[0].Values[1].(string))

			list := []string{}

			// 2. get all 'function_start'
			// generate a value for the 'exported_instance' values, e.g. "935882-0-A000-00002" ==> "935882-0-A000-.*"
			exportedInstance := getExportedInstanceValues(execID, job_id)

			// create query string. Example: function_start{function_name="test_function_1",exported_instance=~"935882-0-.*"}
			queryStr = "function_start{function_name='" + resource + "',exported_instance=~'" + exportedInstance + "'}"
			log.Infof("guarantees [notStarted] > Getting all 'function_start' values from Prometheus > queryStr: %s", queryStr)

			// execut query
			resQ, err = Query(queryStr, prometheusRootURL)
			if err == nil && len(resQ.Data.Results) > 0 {
				list = processQuery(resQ, jobTotalCalls, execID, job_id) // return []string{}
			} else if err == nil && len(resQ.Data.Results) == 0 {
				log.Warnf("guarantees [notStarted] > query with no results = 'function_start' not found for defined exportedInstance values")
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
func getExportedInstanceValues(execID string, job_id string) string {
	return execID + "-" + job_id + "-.*"
}

// processQuery process the results and returns a list of not started exportedInstances
func processQuery(q QueryResp, t int, execID string, job_id string) []string {
	list := []string{}

	for i := 0; i < t; i++ {
		exportedInstanceValue := getExportedInstanceValue(execID, job_id, generateCallId(i))
		found := false

		for _, r := range q.Data.Results {
			if r.Metric.ExportedInstance == exportedInstanceValue {
				found = true
				break
			}
		}

		if !found {
			list = append(list, exportedInstanceValue)
		}
	}

	return list
}

// generates a call_id
func generateCallId(i int) string {
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

// getExportedInstanceValue
func getExportedInstanceValue(execID string, job_id string, callId string) string {
	return execID + "-" + job_id + "-" + callId
}

/*
CheckGuarantee checks the guarantee

func CheckGuarantee(prometheusRootURL string, execID string, resource string) []string {
	return notStarted(prometheusRootURL, execID, resource)
}
*/
