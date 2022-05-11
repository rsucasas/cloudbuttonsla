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

package utils

import (
	"encoding/json"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"

	"SLALite/model"
)

/*
This file manages the default templates used by the SLA. These templates are
located in "resources/templates"
*/

type TEMPLATES_LIST []model.Template

var TemplatesList TEMPLATES_LIST

// loadTemplates
func loadTemplates(r model.IRepository) {
	log.Println("Templates [loadTemplates] Looking for default templates declared in '/resources/templates/templates.json' ...")

	slaTemplatesPath := "./resources/templates/templates.json"
	log.Println("Templates [loadTemplates] Reading content of templates.json file [" + slaTemplatesPath + "] ...")

	file, err := os.Open(slaTemplatesPath)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&TemplatesList)
	if err != nil {
		panic(err)
	}

	log.Println("Templates [loadTemplates] List of available Templates ...")
	for i := range TemplatesList {
		log.Println("Templates [loadTemplates] > [" + strconv.Itoa(i) + "] " + TemplatesList[i].Name)

		_, err = r.CreateTemplate(&TemplatesList[i])
		if err != nil {
			log.Error("Templates [loadTemplates] > ["+strconv.Itoa(i)+"] Error saving '"+TemplatesList[i].Name+"': ", err)
		} else {
			log.Println("Templates [loadTemplates] > [" + strconv.Itoa(i) + "] Template '" + TemplatesList[i].Name + "' added to Repository.")
		}
	}
}

/*
LoadDefaultTemplates Load default templates from templates folder
*/
func LoadDefaultTemplates(r model.IRepository) {
	log.Println("Templates [LoadDefaultTemplates] Loading default templates ...")
	loadTemplates(r)
}
