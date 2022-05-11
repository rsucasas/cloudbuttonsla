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
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"
)

type query struct {
	Status string `json:"status"`
	Data   data   `json:"data"`
}

type resultType string

type data struct {
	ResultType resultType `json:"resultType"`
	Results    []result   `json:"result"`
}

type result struct {
	Metric metric  `json:"metric"`
	Item   value   `json:"value"`
	Items  []value `json:"values"`
}

type metric struct {
	Name string `json:"__name__"`
	//Instance         string `json:"instance"`
	ExportedInstance string `json:"exported_instance"`
	Job              string `json:"job_id"`
	Function         string `json:"function_name"`
	Call             string `json:"call_id"`
}

type value struct {
	Timestamp datetime
	Value     float64
}

func (v *value) UnmarshalJSON(data []byte) error {
	var sv string
	aux := [...]interface{}{&v.Timestamp, &sv}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	n, err := strconv.ParseFloat(sv, 64)
	if err != nil {
		return err
	}
	v.Value = n
	return nil
}

type datetime time.Time

func (t *datetime) UnmarshalJSON(data []byte) error {
	/*
	 * From a ssss[.nnnnn] string, parses secons a nanoseconds separately
	 * To calculate nsecs, nnnnn is padded right with zeros as to have 9 digits.
	 */
	s := string(data)

	//	aux := strings.Split(s, ".")
	var parts [2]string
	copy(parts[:], strings.Split(s, "."))

	secs, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return err
	}

	parts[1] = padzerosright(parts[1])

	nsecs, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}

	*t = datetime(time.Unix(secs, nsecs))
	return nil
}

func padzerosright(i string) string {
	if len(i) >= 9 {
		return i[0:9]
	}
	zeros := "000000000"
	o := i + zeros[0:len(zeros)-len(i)]
	return o
}

// GetTime converts to a time from a unix float number of seconds.
// This loses precision. E.g.
// getTime(1571988825.63) = 2019-10-25 07:33:45.630000128
func getTime(secs float64) time.Time {
	ns := int64(math.Round(secs * 1000000000))
	return time.Unix(0, ns)
}
