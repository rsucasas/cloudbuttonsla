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
package prometheus

import (
	"SLALite/assessment"
	"SLALite/assessment/monitor/genericadapter"
	"SLALite/utils"
	"fmt"
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestRetrieve(t *testing.T) {

	prometheusURL, ok := os.LookupEnv("SLA_PROMETHEUSURL")

	if !ok {
		log.Info("Skipping integration test")
		t.SkipNow()
	}

	retr := Retriever{
		URL: prometheusURL,
	}
	a, _ := utils.ReadAgreement("testdata/a.json")
	f := retr.Retrieve()
	now := time.Now()
	items := assessment.BuildRetrievalItems(&a, a.Details.Guarantees[0], []string{"execution_time"}, now)
	metrics := f(a, items)

	fmt.Printf("values=%v\n", metrics)
}

func TestRetrieve2(t *testing.T) {
	// this test assumes VideoIntelligence sample data on Prometheus

	prometheusURL, ok := os.LookupEnv("SLA_PROMETHEUSURL")

	if !ok {
		log.Info("Skipping integration test")
		t.SkipNow()
	}

	retr := Retriever{
		URL: prometheusURL,
	}
	a, _ := utils.ReadAgreement("testdata/b.json")

	adapter := genericadapter.New(retr.Retrieve(), genericadapter.Identity)
	now := time.Date(2019, 10, 29, 12, 5, 0, 0, time.Local)

	cfg := assessment.Config{
		Adapter: adapter,
		Now:     now,
	}
	result := assessment.AssessAgreement(&a, cfg)
	fmt.Printf("%#v", result)
}
