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

/*
Package mongodb is an implementation of a model.IRepository backed up by a mongodb.
*/
package mongodb

import (
	"SLALite/model"

	log "github.com/sirupsen/logrus"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/spf13/viper"
)

const (
	// Name is the unique identifier of this repository
	Name                    string = "mongodb"
	defaultURL              string = "localhost"
	repositoryDbName        string = "slalite"
	providersCollectionName string = "Providers"
	agreementCollectionName string = "Agreements"
	templateCollectionName  string = "Templates"
	violationCollectionName string = "Violations"

	mongoConfigName string = "mongodb.yml"

	connectionURL string = "connection"
	mongoDatabase string = "database"
	clearOnBoot   string = "clear_on_boot"
)

//Repository contains the repository persistence implementation based on MongoDB
type Repository struct {
	session  *mgo.Session
	database *mgo.Database
}

// NewDefaultConfig gets a default configuration for a Repository
func NewDefaultConfig() (*viper.Viper, error) {
	config := viper.New()

	config.SetEnvPrefix("sla") // Env vars start with 'SLA_'
	config.AutomaticEnv()
	config.SetConfigName(mongoConfigName)
	config.AddConfigPath(model.UnixConfigPath)
	setDefaults(config)

	confError := config.ReadInConfig()
	if confError != nil {
		log.Println("Can't find MongoDB configuration file: " + confError.Error())
		log.Println("Using defaults")
	}

	return config, confError
}

func setDefaults(config *viper.Viper) {
	config.SetDefault(connectionURL, defaultURL)
	config.SetDefault(mongoDatabase, repositoryDbName)
	config.SetDefault(clearOnBoot, false)
}

// New creates a new instance of the Repository with the database configurarion read from a configuration file
func New(config *viper.Viper) (Repository, error) {
	if config == nil {
		config, _ = NewDefaultConfig()
	} else {
		setDefaults(config)
	}

	logConfig(config)

	repo := new(Repository)

	session, err := mgo.Dial(config.GetString(connectionURL))
	if err != nil {
		log.Fatal("Error getting connection to Mongo DB: " + err.Error())
	}

	database := session.DB(config.GetString(mongoDatabase))
	clear := config.GetBool(clearOnBoot)
	if clear {
		err := database.DropDatabase()
		if err != nil {
			log.Println("Error dropping database " + repositoryDbName + ": " + err.Error())
		}
	}

	repo.session = session
	repo.database = database

	return *repo, err
}

func logConfig(config *viper.Viper) {
	log.Printf("MongoDB configuration\n"+
		"\tconnectionURL: %v\n"+
		"\tdatabaseName: %v\n"+
		"\tclear on boot: %v\n",
		config.GetString(connectionURL),
		config.GetString(mongoDatabase), config.GetBool(clearOnBoot))
}

func (r Repository) getList(collection string, query, result interface{}) (interface{}, error) {
	err := r.database.C(collection).Find(query).All(result)
	return result, err
}

func (r Repository) getAll(collection string, result interface{}) (interface{}, error) {
	return r.getList(collection, bson.M{}, result)
}

func (r Repository) get(collection string, id string, result model.Identity) (model.Identity, error) {
	err := r.database.C(collection).FindId(id).One(result)
	if err == mgo.ErrNotFound {
		return result, model.ErrNotFound
	}

	return result, err
}

func (r Repository) create(collection string, object model.Identity) (model.Identity, error) {
	_, err := r.get(collection, object.GetId(), object)
	if err != model.ErrNotFound {
		return object, model.ErrAlreadyExist
	}
	errCreate := r.database.C(collection).Insert(object)
	return object, errCreate
}

func (r Repository) update(collection, id string, upd interface{}) error {
	err := r.database.C(collection).UpdateId(id, upd)
	if err == mgo.ErrNotFound {
		return model.ErrNotFound
	}
	return err
}

func (r Repository) delete(collection, id string) error {
	error := r.database.C(collection).RemoveId(id)
	if error == mgo.ErrNotFound {
		return model.ErrNotFound
	}
	return error

}

/*
GetAllProviders returns the list of providers.

The list is empty when there are no providers;
error != nil on error
*/
func (r Repository) GetAllProviders() (model.Providers, error) {
	res, err := r.getAll(providersCollectionName, new(model.Providers))
	return *((res).(*model.Providers)), err
}

/*
GetProvider returns the Provider identified by id.

error != nil on error;
error is sql.ErrNoRows if the provider is not found
*/
func (r Repository) GetProvider(id string) (*model.Provider, error) {
	res, err := r.get(providersCollectionName, id, new(model.Provider))
	return res.(*model.Provider), err
}

/*
CreateProvider stores a new provider.

error != nil on error;
error is sql.ErrNoRows if the provider already exists
*/
func (r Repository) CreateProvider(provider *model.Provider) (*model.Provider, error) {
	res, err := r.create(providersCollectionName, provider)
	return res.(*model.Provider), err
}

/*
DeleteProvider deletes from the repository the provider whose id is provider.Id.

error != nil on error;
error is sql.ErrNoRows if the provider does not exist.
*/
func (r Repository) DeleteProvider(provider *model.Provider) error {
	return r.delete(providersCollectionName, provider.Id)
}

/*
GetAllAgreements returns the list of agreements.

The list is empty when there are no agreements;
error != nil on error
*/
func (r Repository) GetAllAgreements() (model.Agreements, error) {
	res, err := r.getAll(agreementCollectionName, new(model.Agreements))
	return *((res).(*model.Agreements)), err
}

/*
GetAgreement returns the Agreement identified by id.

error != nil on error;
error is sql.ErrNoRows if the Agreement is not found
*/
func (r Repository) GetAgreement(id string) (*model.Agreement, error) {
	res, err := r.get(agreementCollectionName, id, new(model.Agreement))
	return res.(*model.Agreement), err
}

/*
GetAgreementsByState returns the agreements that have one of the items in states.

error != nil on error;
*/
func (r Repository) GetAgreementsByState(states ...model.State) (model.Agreements, error) {
	output := new(model.Agreements)

	query := bson.M{"state": bson.M{"$in": states}}
	result, err := r.getList(agreementCollectionName, query, output)
	return *((result).(*model.Agreements)), err
}

/*
CreateAgreement stores a new Agreement.

error != nil on error;
error is sql.ErrNoRows if the Agreement already exists
*/
func (r Repository) CreateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	res, err := r.create(agreementCollectionName, agreement)
	return res.(*model.Agreement), err
}

/*
UpdateAgreement updates the information of an already saved instance of an agreement
*/
func (r Repository) UpdateAgreement(agreement *model.Agreement) (*model.Agreement, error) {
	err := r.update(agreementCollectionName, agreement.Id, agreement)
	return agreement, err
}

/*
DeleteAgreement deletes from the repository the Agreement whose id is provider.Id.

error != nil on error;
error is sql.ErrNoRows if the Agreement does not exist.
*/
func (r Repository) DeleteAgreement(agreement *model.Agreement) error {
	return r.delete(agreementCollectionName, agreement.Id)
}

/*
CreateViolation stores a new Violation.

error != nil on error;
error is sql.ErrNoRows if the Violation already exists
*/
func (r Repository) CreateViolation(v *model.Violation) (*model.Violation, error) {
	res, err := r.create(violationCollectionName, v)
	return res.(*model.Violation), err
}

/*
GetViolation returns the Violation identified by id.

error != nil on error;
error is sql.ErrNoRows if the Violation is not found
*/
func (r Repository) GetViolation(id string) (*model.Violation, error) {
	res, err := r.get(violationCollectionName, id, new(model.Violation))
	return res.(*model.Violation), err
}

/*
UpdateAgreementState transits the state of the agreement
*/
func (r Repository) UpdateAgreementState(id string, newState model.State) (*model.Agreement, error) {

	var err error
	var agreement *model.Agreement

	err = r.update(agreementCollectionName, id, bson.M{"$set": bson.M{"state": newState}})
	if err == nil {
		agreement, _ = r.GetAgreement(id)
	}
	return agreement, err
}

/*
GetAllTemplates returns the list of templates.

The list is empty when there are no templates;
error != nil on error
*/
func (r Repository) GetAllTemplates() (model.Templates, error) {
	res, err := r.getAll(templateCollectionName, new(model.Templates))
	return *((res).(*model.Templates)), err
}

/*
GetTemplate returns the Template identified by id.

error != nil on error;
error is sql.ErrNoRows if the Template is not found
*/
func (r Repository) GetTemplate(id string) (*model.Template, error) {
	res, err := r.get(templateCollectionName, id, new(model.Template))
	return res.(*model.Template), err
}

/*
CreateTemplate stores a new Template.

error != nil on error;
error is sql.ErrNoRows if the Template already exists
*/
func (r Repository) CreateTemplate(template *model.Template) (*model.Template, error) {
	res, err := r.create(templateCollectionName, template)
	return res.(*model.Template), err
}
