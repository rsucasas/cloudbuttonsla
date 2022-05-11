# cloudbutton-SLA #

## Description ##

cloudbutton-SLA is an SLA system, inspired by the WS-Agreement standard, that uses Knative Observability Plugin monitoring data to supervise Knative running pods in order to identify candidate performance improvements and/or problems and inform on a Rabbit queue of violations on the agreements that made up the SLA. 

Its features are:
* REST interface to manage creation and update of agreements
* Agreements evaluation on background; any breach in the agreement terms generates an SLA violation.
* Configurable monitoring: a monitoring has to be provided externally. For CloudButton project, the monitoring system is Prometheus, but others are also available.
* Configurable repository: a memory repository (for developing purposes) and a mongodb repository are provided, but more can be added.
* Configurable notifier: violations on the SLA information are made available from different notifiers. For CloudButton project, Rabbit has been chosen.

An agreement is represented by a simple JSON structure 
(see examples in resources/samples):

```
{
    "id": "a4",
    "name": "an-agreement-name",
    "state": "started",
    "details":{
        "id": "a4",
        "type": "agreement",
        "name": "an-agreement-name",
        "provider": { "id": "a-provider", "name": "A provider" },
        "client": { "id": "a-client", "name": "A client" },
        "creation": "2020-01-01T17:09:45Z",
        "expiration": "2021-01-01T17:09:45Z",
        "variables": [
            {
                "name": "reconciler",
                "metric": "sum%20by%20(reconciler)(60*rate(controller_reconcile_count[1m]))"
            }
        ],
        "guarantees": [
            {
                "name": "Reconciler Less than 10",
                "constraint": "reconciler < 10"
            }
        ]
    }
}

```

## Quick usage guide ##

### Installation ###

Build the Docker image:

    make docker

Run the container:

    docker run -ti -p 8090:8090 slalite:<version>

Stop execution pressing CTRL-C

To run the service under HTTPs, you must change supply a different configuration file and the certificate files. You will find these files in docker/https for debugging purposes. DO NOT USE THE CERT.PEM and KEY.PEM in production!!

    docker run -ti -p 8090:8090 -v $PWD/docker/https:/etc/slalite slalite

### Configuration ###

cloudbutton-SLA can be configured with a configuration file and with environment 
variables. The configuration file is read by default from /etc/slalite and the current 
working directory. The `-f` parameter can be used to set the config file location.

```
$ ./SLALite -h
Usage of SLALite:
  -b string
        Filename (w/o extension) of config file (default "slalite")
  -d string
        Directories where to search config files (default "/etc/slalite:.")
  -f string
        Path of configuration file. Overrides -b and -d
```

#### File settings ####

*General settings*

* `singlefile` (default: `false`). Sets if all file settings are read 
  from a single file or from several files. For example, when `singlefile=false`,
  the MongoDB settings are read from the file `mongodb.yml`.
* `repository` (default: `memory`). Sets the repository type to use. Set this
  value to `mongodb` to use a MongoDB database.
* `externalIDs` (default: `false`). Set this to true if the repository auto assign 
  the IDs of the saved entities.
* `checkPeriod` (default: `60s`). Sets the period of assessments executions, in the
  format of a time.Duration (e.g. 60s, 1.5m). If no unit is given, seconds are assumed.
* `transientTime` (default: `0s`). Sets the transient time after a violation on a 
  guarantee term is raised, in the format of a time.Duration (e.g. 60s, 1.5m). No more 
  violations on that term will be raised while in the transient time. 
  If no unit is given, seconds are assumed.
* `CAPath`. Sets the value of a file path containing certificates of trusted
  CAs; to be used to connect as client to SSL servers whose certificate is
  not trusted by default (e.g. self-signed certificates)

*REST interface settings*

* `port` (default: `8090`). Port of REST interface.
* `enableSsl` (default: `false`). Enables the use of SSL on the REST
  interface. The two following variables should be set.
* `sslCertPath` (default: `cert.pem`). Sets the certificate path.
* `sslKeyPath` (default: `key.pem`). Sets the private key path to access the
  certificate.

*MongoDB settings (default file: /etc/slalite/mongodb.yml)*

* `connection` (default: `localhost`). Sets the MongoDB host.
* `database` (default: `slalite`). Sets the MongoDB database name to use.
* `clear_on_boot` (default: `false`). Sets if the database is cleared on
  startup (useful for tests).

#### Env vars  ####

Every file setting can be overriden with the use of environment variables.
The name of the var is the uppercase setting name prefixed with `SLA_`. For
example, to override the check period, set the env var `SLA_CHECKPERIOD`.

### Usage ###

cloudbutton-SLA offers a usual REST API, with an endpoint on /agreements

Add an agreement (agreement below is stopped):

    curl -k -X POST -d @resources/samples/agreement.json http://localhost:8090/agreements
    

Change agreement state:

    curl -k http://localhost:8090/agreements/a02 -X PATCH -d'{"state":"started"}'

Get agreements:

    curl -k http://localhost:8090/agreements
    curl -k http://localhost:8090/agreements/a02

Add a template:

    curl -k -X POST -d @resources/samples/template.json http://localhost:8090/templates

Get templates:

    curl -k http://localhost:8090/templates
    curl -k http://localhost:8090/templates/t01

Create agreement from template:

    curl -k -X POST -d @resources/samples/create-agreement.json http://localhost:8090/create-agreement

    curl -X POST -d @agreement_template_cloudbutton.json http://localhost:8091/create-agreement

    {"template_id":"t01","agreement_id":"9be511e8-347f-4a40-b784-e80789e4c65b","parameters":{"M":1,"N":100,"agreementname":"An agreement name","client":{"id":"client01","name":"A name of a client"},"provider":{"id":"provider01","name":"A name of a provider"}}}