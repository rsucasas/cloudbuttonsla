/*
Copyright 2019 Atos

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
	"SLALite/model"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestNew(t *testing.T) {
	config := viper.New()

	config.SetEnvPrefix("sla") // Env vars start with 'SLA_'
	config.AutomaticEnv()
	config.Set(PrometheusURLPropertyName, defaultURL)

	New(config)
}

func TestParseVector(t *testing.T) {
	query, err := readFile("testdata/vector.json")
	if err != nil {
		t.Fatal(err)
	}

	if query.Data.ResultType != vectorType {
		t.Fatalf("Expected: %s; Actual: %s", vectorType, query.Data.ResultType)
	}

	if expected, actual := 1, len(query.Data.Results); actual != expected {
		t.Fatalf("Expected: %d; Actual: %d", expected, actual)
	}

	result := query.Data.Results[0]

	if expected, actual := "go_memstats_frees_total", result.Metric.Name; expected != actual {
		t.Fatalf("Expected: %s; Actual: %s", expected, actual)
	}

	if expected, actual := 862037.0, result.Item.Value; actual != expected {
		t.Fatalf("Expected: %f; Actual: %f", expected, actual)
	}

	if expected, actual := int64(157198882563*1e7), time.Time(result.Item.Timestamp).UnixNano(); expected != actual {
		t.Fatalf("Expected: %d; Actual: %d", expected, actual)
	}
}

func TestParseVectorSeveralMetrics(t *testing.T) {
	query, err := readFile("testdata/vector2.json")
	if err != nil {
		t.Fatal(err)
	}

	if query.Data.ResultType != vectorType {
		t.Fatalf("Expected: %s; Actual: %s", vectorType, query.Data.ResultType)
	}

	if expected, actual := 3, len(query.Data.Results); actual != expected {
		t.Fatalf("Expected: %d; Actual: %d", expected, actual)
	}
}
func TestParseMatrix(t *testing.T) {
	query, err := readFile("testdata/matrix.json")
	if err != nil {
		t.Fatal(err)
	}

	if query.Data.ResultType != matrixType {
		t.Fatalf("Expected: %s; Actual: %s", vectorType, query.Data.ResultType)
	}

	if expected, actual := 2, len(query.Data.Results); actual != expected {
		t.Fatalf("Expected: %d; Actual: %d", expected, actual)
	}

	result := query.Data.Results[0]

	if expected, actual := "prometheus_http_response_size_bytes_count", result.Metric.Name; expected != actual {
		t.Fatalf("Expected: %s; Actual: %s", expected, actual)
	}

	if expected, actual := 2, len(result.Items); actual != expected {
		t.Fatalf("Expected: %d; Actual: %d", expected, actual)
	}

	if expected, actual := int64(1572339600), time.Time(result.Items[0].Timestamp).Unix(); actual != expected {
		t.Fatalf("Expected: %d; Actual: %d", expected, actual)
	}

	if expected, actual := float64(3), result.Items[0].Value; actual != expected {
		t.Fatalf("Expected: %f; Actual: %f", expected, actual)
	}
}

func TestPrometheusRoot(t *testing.T) {
	a1 := model.Agreement{}
	a2 := model.Agreement{
		Assessment: model.Assessment{
			MonitoringURL: "http://localhost:8080",
		},
	}
	r := Retriever{
		URL: "http://localhost:9090",
	}
	if expected := r.URL; r.prometheusRoot(a1) != r.URL {
		t.Errorf("Expected: %s; Actual: %s", expected, r.prometheusRoot(a1))
	}
	if expected := a2.Assessment.MonitoringURL; r.prometheusRoot(a2) != expected {
		t.Errorf("Expected: %s; Actual: %s", expected, r.prometheusRoot(a2))
	}

}

func readFile(path string) (query, error) {
	var result query

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return result, err
	}
	err = parse(f, &result)
	return result, err
}
