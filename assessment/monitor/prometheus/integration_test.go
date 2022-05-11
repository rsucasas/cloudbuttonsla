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
