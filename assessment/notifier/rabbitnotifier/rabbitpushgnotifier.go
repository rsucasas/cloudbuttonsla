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

// Package rabbitpushgnotifier contains a simple ViolationsNotifier that send violations to a rabbit queue.
package rabbitpushgnotifier

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	assessment_model "SLALite/assessment/model"
	"SLALite/assessment/notifier"
	"SLALite/model"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

const (
	// PushgatewayURLPropertyName is the config property name of the Pushgateway URL
	PushgatewayURLPropertyName = "pushgatewayUrl"
	// RabbitMQPropertyName is the config property name of the Rabbit connection
	RabbitMQPropertyName = "rabbitMQ"
	// Name is the unique identifier of this notifier
	Name = "rabbitpushg"
	// PrometheusURLPropertyName is the config property name of the Prometheus URL
	PrometheusURLPropertyName = "prometheusUrl"

	// defaultpushgURL is the value of the Pushgateway URL if PushgatewayURLPropertyName is not set
	defaultpushgURL = "http://77.231.202.2:30091"
	// defaultrabbitMQ is the value of the Rabbit Amqp if RabbitAmqpPropertyName is not set
	defaultrabbitMQ = "amqp://cloudbutton:CloudButton2020@77.231.202.2:5672/"
	// defaultPrometheusURL is the value of the Prometheus URL if PrometheusURLPropertyName is not set
	defaultPrometheusURL = "http://77.231.202.2:30990/"
)

var q amqp.Queue
var ch *amqp.Channel

//var conn *amqp.Connection

//RabbitpushgNotifier logs violations on a Rabbit queue and on a Pushgateway monitor.
type RabbitpushgNotifier struct {
	rabbitMQ      string
	pushgURL      string
	prometheusURL string
}

// New constructs a Rabbitpushg Notifier
func New(config *viper.Viper) notifier.ViolationNotifier {
	config.SetDefault(RabbitMQPropertyName, defaultrabbitMQ)
	config.SetDefault(PushgatewayURLPropertyName, defaultpushgURL)
	config.SetDefault(PrometheusURLPropertyName, defaultPrometheusURL)

	logConfig(config)

	return _new(config.GetString(RabbitMQPropertyName), config.GetString(PushgatewayURLPropertyName), config.GetString(PrometheusURLPropertyName))
}

func logConfig(config *viper.Viper) {
	log.Infof("Rabbit and Pushgateway Notifier configuration\n"+
		"\tRabbitMQ Connection: %s\n\tPushgateway URL: %s\n\tPrometheusURL URL: %s\n",
		config.GetString(RabbitMQPropertyName),
		config.GetString(PushgatewayURLPropertyName),
		config.GetString(PrometheusURLPropertyName))
}

func _new(rabbitmq string, pushgurl string, prometheusURL string) notifier.ViolationNotifier {
	return RabbitpushgNotifier{
		rabbitMQ:      rabbitmq,
		pushgURL:      pushgurl,
		prometheusURL: prometheusURL,
	}

}

// get prometheus root URL
func prometheusRoot(agreement model.Agreement, prometheusURL string) string {
	if agreement.Assessment.MonitoringURL != "" {
		return agreement.Assessment.MonitoringURL
	}
	return prometheusURL
}

// NotifyViolations implements ViolationNotifier interface
func (n RabbitpushgNotifier) NotifyViolations(agreement *model.Agreement, result *assessment_model.Result) {
	prometheusRootURL := prometheusRoot(*agreement, n.prometheusURL) // prometheus URL

	// connection to RabbitMQ
	errRabbitMQ := ConnectQueue(n.rabbitMQ)

	// Notify violations
	log.Debug("Notifying violations of agreement [" + agreement.Id + "] ...")
	funclistlen := 0
	for i, v := range result.Violated {
		if len(v.Violations) > 0 {
			log.Debug("Failed guarantee: " + i)
			for _, vi := range v.Violations {
				log.Infof("Failed guarantee %v of agreement %s at %s with values %s", vi.Guarantee, vi.AgreementId, vi.Datetime, vi.Values)

				// Send to RabbitMQ
				if errRabbitMQ == nil {
					log.Debug("Sending violation to RabbitMQ ...")
					// RabbitMQ Body
					fields := make(map[string]interface{})
					fields["AgreementId"] = vi.AgreementId
					fields["Guarantee"] = vi.Guarantee
					fields["ViolationTime"] = vi.Datetime

					body := make(map[string]interface{})
					body["Message"] = vi.Values
					body["Fields"] = fields

					// handle special cases
					if strings.ToLower(vi.Guarantee) == "notstarted" {
						// "notstarted"
						log.Debugf("Processing 'notstarted' violation ...")

						var funcList []string = ChecknotStartedFunctions(prometheusRootURL, vi.Values[0].Key, vi.Values[0].Resource)
						funclistlen = len(funcList)
						fields["notStarted"] = funcList
						log.Debugf("List of not started functions: %v", funcList)

						// send one message to Kafka for each of the not started functions found
						if funclistlen > 0 {
							for _, v := range funcList {
								log.Infof("Sending function violation [" + v + "] to RabbitMQ queue [" + q.Name + "] ... ")
								sendNotStartedFunctionViolation(v, vi.AgreementId, vi.Guarantee, vi.Datetime, vi.Values)
							}
						}
					} else if strings.Contains(strings.ToLower(vi.Guarantee), "toocostly") {
						// "toocostly"
						log.Debugf("Processing 'toocostly' violation ...")

					}

					jsonData, err := json.Marshal(body)
					failOnError(err, "Failed to Marshal body")

					log.Infof("Sending list of functions violation to RabbitMQ queue [" + q.Name + "] ... ")
					fmt.Println(string(jsonData))

					err = ch.Publish(
						"",     // exchange
						q.Name, // routing key
						false,  // mandatory
						false,  // immediate
						amqp.Publishing{
							ContentType: "application/json",
							Body:        jsonData,
						})
					failOnError(err, "Failed to publish a message")
					log.Infof("Violation Message Published on queue %s", q.Name)
				}

				// Send information to Prometheus Pushgateway to record violation in Prometheus
				// 	- instance: 		vi.Values[0].Key
				// 	- function_name: 	vi.Values[0].Resource
				//  - agreement and guarantee
				//  - time of first violation: vi.Values[0].DateTime

				log.Debug("Sending violation to Prometheus Pushgateway ...")
				SendViolationToPrometheus(n.pushgURL, vi.AgreementId, vi.Guarantee, vi.Values[0].Key,
					vi.Values[0].DateTime.String(), vi.Values[0].Resource, funclistlen)

				//RemoveMetric(n.pushgURL, vi.Values[k].Key, vi.Values[k].Resource)
			}
		}
	}
}

// ConnectQueue connect to Rabbit queue
func ConnectQueue(rabURL string) error {
	log.Debugf("Connecting to RabbitMQ [" + rabURL + "] ... ")

	conn, err := amqp.Dial(rabURL)
	if err != nil {
		log.Error("Failed to connect to RabbitMQ: ", err)
	} else {
		//defer conn.Close()

		ch, err = conn.Channel()
		if err != nil {
			log.Error("Failed to open a channel: ", err)
		} else {
			//defer ch.Close()

			q, err = ch.QueueDeclare(
				"CloudButton", // name
				true,          // durable
				false,         // delete when unused
				false,         // exclusive
				false,         // no-wait
				nil,           // arguments
			)
			if err != nil {
				log.Error("Failed to declare a queue: ", err)
			}
		}
	}

	return err
}

// sendNotStartedFunctionViolation sends a not started function violation message to Rabbit
func sendNotStartedFunctionViolation(v string, a string, g string, t time.Time, vs []model.MetricValue) {
	fields := make(map[string]interface{})
	fields["AgreementId"] = a
	fields["Guarantee"] = g
	fields["ViolationTime"] = t

	body := make(map[string]interface{})
	for i, _ := range vs {
		vs[i].Key = v
	}
	body["Message"] = vs
	body["Fields"] = fields

	fields["notStarted"] = v

	jsonData, err := json.Marshal(body)
	failOnError(err, "Failed to Marshal body")

	log.Infof("[sendNotStartedFunctionViolation] jsonData content:")
	fmt.Println(string(jsonData))

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		})
	failOnError(err, "Failed to publish a message")
	log.Infof("Violation Message Published on queue %s", q.Name)
}

// SendViolationToPrometheus violations to prometheus via pushgateway
func SendViolationToPrometheus(pushg string, agr string, gua string, key string, violtime string, fun string, listlen int) {
	violationValue := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "CloudButton",
		Subsystem:   "sla",
		Name:        "QoS_violation",
		Help:        "Date of last violation.",
		ConstLabels: map[string]string{},
	})

	violationValue.SetToCurrentTime()
	if listlen > 0 {
		violationValue.Set(float64(listlen))
	}
	var na = "NA"
	if fun == "" {
		fun = na
	}

	if err := push.New(pushg, "sla").
		Collector(violationValue).
		Grouping("agreement", agr).
		Grouping("guarantee", gua).
		Grouping("function_name", fun).
		Grouping("violation_time", violtime).
		Grouping("instance", key).
		Push(); err != nil {
		log.Error("Could not push violation time to Pushgateway: " + err.Error())
	}
}

/*
//RemoveMetric Remove the metric that caused a violation from PushGateway, after saving it in Prometheus
func RemoveMetric(pushg string, delKey string, delRes string) {
	log.Infof("Key: %s , Resource: %s", delKey, delRes)
	split := strings.Split(delKey, "-")
	shortKey := split[0]
	executorID := split[1]
	jobID := split[2]
	callID := split[3]
	log.Infof("shortKey: %s , executorID: %s, jobID: %s , callID: %s", shortKey, executorID, jobID, callID)

	routeURL := pushg + "/metrics/job/lithops/call_id/" + callID + "/function_name/" + delRes + "/instance/" + delKey + "/job_id/" + jobID
	log.Infof("routeURL: %s", routeURL)
	/*res, err := http.NewRequest("DELETE", routeURL, nil)
	if err != nil {
		log.Error("Error: " + err.Error())
	}
	defer res.Body.Close()
	message, _ := ioutil.ReadAll(res.Body)
	log.Debug("Response: " + string(message))
}
*/

//Error traitment
func failOnError(err error, msg string) {
	if err != nil {
		log.Error("%s: %s", msg, err)
	}
}
