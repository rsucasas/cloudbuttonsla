package rest

import (
	amodel "SLALite/assessment/model"
	"SLALite/assessment/notifier"
	"SLALite/model"
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

const (
	// NotificationURLPropertyName is the name of the property notificationUrl
	NotificationURLPropertyName = "notificationUrl"
	// Name is the unique identifier of this notifier
	Name = "rest"
)

type _notifier struct {
	url string
}

type violationInfo struct {
	Type          string            `json:"type"`
	AgreementID   string            `json:"agremeent_id"`
	Client        model.Client      `json:"client"`
	GuaranteeName string            `json:"guarantee_name"`
	Violations    []model.Violation `json:"violations"`
}

// New constructs a REST Notifier
func New(config *viper.Viper) notifier.ViolationNotifier {

	logConfig(config)
	return _new(config.GetString(NotificationURLPropertyName))
}

func _new(url string) notifier.ViolationNotifier {
	return _notifier{
		url: url,
	}
}

func logConfig(config *viper.Viper) {
	log.Printf("RestNotifier configuration\n"+
		"\tURL: %v\n",
		config.GetString(NotificationURLPropertyName))

}

/* Implements notifier.NotifyViolations */
func (not _notifier) NotifyViolations(agreement *model.Agreement, result *amodel.Result) {

	vs := result.GetViolations()
	if len(vs) == 0 {
		return
	}

	info := violationInfo{
		Type:        "violation",
		AgreementID: agreement.Id,
		Client:      agreement.Details.Client,
		Violations:  vs,
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(info)

	_, err := http.Post(not.url, "application/json; charset=utf-8", b)

	if err != nil {
		log.Errorf("RestNotifier error: %s", err)
	} else {
		log.Infof("RestNotifier. Sent violations: %v", info)
	}
}
