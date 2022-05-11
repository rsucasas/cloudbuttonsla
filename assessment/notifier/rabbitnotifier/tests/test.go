/*
Copyright 2017 Atos

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
package main

import (
	rabbitpushgnotifier "SLALite/assessment/notifier/rabbitnotifier"

	"github.com/prometheus/common/log"
)

func main() {
	rq, err := rabbitpushgnotifier.Query("job_total_calls{function_name='test_function_1',executor_id='935882-0'}", "http://77.231.202.2:30001")
	if err != nil {
		log.Errorf("Error in call to Prometheus %s", err.Error())
	}

	log.Infof("Status: %s", rq.Status)
	log.Infof("Values[0]: %v", rq.Data.Results[0].Values[0])
	log.Infof("Values[1]: %s", rq.Data.Results[0].Values[1].(string))
}
