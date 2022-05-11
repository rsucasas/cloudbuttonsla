package rest

import (
	"SLALite/assessment"
	amodel "SLALite/assessment/model"
	"SLALite/assessment/monitor/simpleadapter"
	"SLALite/model"
	"SLALite/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

/*
To run this test, set up a server that accepts notification requests
and set env var SLA_NOTIFICATION_URL=<server url>.
*/

var agreement model.Agreement
var ma *simpleadapter.ArrayMonitoringAdapter

func TestNew(t *testing.T) {
	config := viper.New()

	config.SetEnvPrefix("sla") // Env vars start with 'SLA_'
	config.AutomaticEnv()
	config.Set(NotificationURLPropertyName, "http://localhost:8080")

	New(config)
}
func TestSend(t *testing.T) {

	Init()
	cfg := assessment.Config{Adapter: ma, Now: time.Now()}
	result, _ := assessment.EvaluateAgreement(&agreement, cfg)
	server := httptest.NewUnstartedServer(http.HandlerFunc(f))
	server.Start()
	defer server.Close()

	not := _new(server.URL)
	not.NotifyViolations(&agreement, &result)
}

func TestSendEmpty(t *testing.T) {

	Init()
	server := httptest.NewUnstartedServer(http.HandlerFunc(f))
	server.Start()
	defer server.Close()

	not := _new(server.URL)
	not.NotifyViolations(&agreement, &amodel.Result{})
}

func TestSendWrong(t *testing.T) {
	Init()
	cfg := assessment.Config{Adapter: ma, Now: time.Now()}
	result, _ := assessment.EvaluateAgreement(&agreement, cfg)
	server := httptest.NewUnstartedServer(http.HandlerFunc(g))
	server.Start()
	defer server.Close()

	not := _new("http://localhost:1")
	not.NotifyViolations(&agreement, &result)
}

func TestSendIntegration(t *testing.T) {
	url, ok := os.LookupEnv("SLA_NOTIFICATION_URL")

	if !ok {
		t.Skip("Skipping integration test")
	}
	cfg := assessment.Config{Adapter: ma, Now: time.Now()}
	result, _ := assessment.EvaluateAgreement(&agreement, cfg)

	not := _new(url)
	not.NotifyViolations(&agreement, &result)
}

func Init() {
	agreement, _ = utils.ReadAgreement("testdata/agreement.json")
	ma = simpleadapter.New(amodel.GuaranteeData{
		amodel.ExpressionData{
			"execution_time": model.MetricValue{
				Key:      "execution_time",
				Value:    1000,
				DateTime: time.Now(),
			},
		},
	})
}

func f(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func g(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not Found"))
}
