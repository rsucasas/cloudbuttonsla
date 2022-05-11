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
	"SLALite/assessment"
	"SLALite/assessment/monitor"
	"SLALite/assessment/monitor/genericadapter"
	"SLALite/assessment/monitor/prometheus"
	"SLALite/assessment/notifier"
	"SLALite/assessment/notifier/lognotifier"
	rabbitpushgnotifier "SLALite/assessment/notifier/rabbitnotifier"
	"SLALite/assessment/notifier/rest"
	"SLALite/model"
	"SLALite/repositories/memrepository"
	"SLALite/repositories/mongodb"
	"SLALite/repositories/validation"
	"SLALite/utils"
	"flag"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

// version and date are defined on compilation (see makefile)
var version string
var date string

func main() {

	// TODO: Add windows path
	configPath := flag.String("d", utils.UnixConfigPath, "Directories where to search config files")
	configBasename := flag.String("b", utils.ConfigName, "Filename (w/o extension) of config file")
	configFile := flag.String("f", "", "Path of configuration file. Overrides -b and -d")
	flag.Parse()

	log.Infof("Running SLALite %s compiled on %s", version, date)
	config := createMainConfig(configFile, configPath, configBasename)
	logMainConfig(config)

	singlefile := config.GetBool(utils.SingleFilePropertyName)
	checkPeriod := asSeconds(config, utils.CheckPeriodPropertyName)
	repoType := config.GetString(utils.RepositoryTypePropertyName)
	trasientTime := asSeconds(config, utils.TransientTimePropertyName)

	utils.AddTrustedCAs(config)

	var repoconfig *viper.Viper
	if singlefile {
		repoconfig = config
	}

	var repo model.IRepository
	var errRepo error
	switch repoType {
	case utils.DefaultRepositoryType:
		repo, errRepo = memrepository.New(repoconfig)
	case "mongodb":
		repo, errRepo = mongodb.New(repoconfig)
	}
	if errRepo != nil {
		log.Fatal("Error creating repository: ", errRepo.Error())
	}

	validator := model.NewDefaultValidator(config.GetBool(utils.ExternalIDsPropertyName), true)

	adapter := buildAdapter(config)

	notifier := buildNotifier(config)

	repo, _ = validation.New(repo, validator)
	if repo != nil {
		a, _ := NewApp(config, repo, validator)
		aCfg := assessment.Config{
			Repo:      repo,
			Adapter:   adapter,
			Notifier:  notifier,
			Transient: trasientTime,
		}
		go createValidationThread(checkPeriod, aCfg)
		a.Run()
	}
}

func buildAdapter(config *viper.Viper) monitor.MonitoringAdapter {
	aType := config.GetString(utils.AdapterTypePropertyName)
	switch aType {
	case prometheus.Name:
		adapter := genericadapter.New(
			prometheus.New(config).Retrieve(),
			genericadapter.Identity)
		return adapter
	default:
		adapter := genericadapter.New(
			genericadapter.DummyRetriever{Size: 3}.Retrieve(),
			genericadapter.Identity)
		return adapter
	}

}

func buildNotifier(config *viper.Viper) notifier.ViolationNotifier {
	nType := config.GetString(utils.NotifierTypePropertyName)
	switch nType {
	case rest.Name:
		notifier := rest.New(config)
		return notifier
	case rabbitpushgnotifier.Name:
		notifier := rabbitpushgnotifier.New(config)
		return notifier
	default:
		notifier := lognotifier.LogNotifier{}
		return notifier
	}

}

func asSeconds(config *viper.Viper, field string) time.Duration {

	raw := config.GetString(field)
	// if it is already a valid duration, return directly
	if _, err := time.ParseDuration(raw); err == nil {
		return config.GetDuration(field)
	}

	// if not, assume it is (decimal) number of seconds; read as ms and convert to seconds.
	ms := config.GetFloat64(field)
	return time.Duration(ms*1000) * time.Millisecond
}

//
// Creates the main Viper configuration.
// file: if set, is the path to a configuration file. If not set, paths and basename will be used
// paths: colon separated paths where to search a config file
// basename: basename of a configuration file accepted by Viper (extension is automatic)
//
func createMainConfig(file *string, paths *string, basename *string) *viper.Viper {
	config := viper.New()

	config.SetEnvPrefix(utils.ConfigPrefix) // Env vars start with 'SLA_'
	config.AutomaticEnv()
	config.SetDefault(utils.CheckPeriodPropertyName, utils.DefaultCheckPeriod)
	config.SetDefault(utils.RepositoryTypePropertyName, utils.DefaultRepositoryType)
	config.SetDefault(utils.AdapterTypePropertyName, utils.DefaultAdapterType)
	config.SetDefault(utils.ExternalIDsPropertyName, utils.DefaultExternalIDs)
	config.SetDefault(utils.TransientTimePropertyName, utils.DefaultTransientTime)

	if *file != "" {
		config.SetConfigFile(*file)
	} else {
		config.SetConfigName(*basename)
		for _, path := range strings.Split(*paths, ":") {
			config.AddConfigPath(path)
		}
	}

	errConfig := config.ReadInConfig()
	if errConfig != nil {
		log.Println("Can't find configuration file: " + errConfig.Error())
		log.Println("Using defaults")
	}
	return config
}

func logMainConfig(config *viper.Viper) {

	checkPeriod := asSeconds(config, utils.CheckPeriodPropertyName)
	repoType := config.GetString(utils.RepositoryTypePropertyName)
	adapterType := config.GetString(utils.AdapterTypePropertyName)
	notifierType := config.GetString(utils.NotifierTypePropertyName)
	externalIDs := config.GetBool(utils.ExternalIDsPropertyName)
	transientTime := asSeconds(config, utils.TransientTimePropertyName)

	log.Infof("SLALite initialization\n"+
		"\tConfigfile: %s\n"+
		"\tRepository type: %s\n"+
		"\tAdapter type: %s\n"+
		"\tNotifier type: %s\n"+
		"\tExternal IDs: %v\n"+
		"\tTransient time: %v\n"+
		"\tCheck period:%v\n",
		config.ConfigFileUsed(), repoType, adapterType, notifierType, externalIDs, transientTime, checkPeriod)

	caPath := config.GetString(utils.CAPathPropertyName)
	if caPath != "" {
		log.Infof("SLALite intialization. Trusted CAs file: %s", caPath)
	}
}

func createValidationThread(checkPeriod time.Duration, cfg assessment.Config) {

	ticker := time.NewTicker(checkPeriod)

	for {
		<-ticker.C
		cfg.Now = time.Now()
		assessment.AssessActiveAgreements(cfg)
	}
}

func validateProviders(repo model.IRepository) {
	providers, err := repo.GetAllProviders()

	if err == nil {
		log.Println("There are " + strconv.Itoa(len(providers)) + " providers")
	} else {
		log.Println("Error: " + err.Error())
	}
}
