[![Build Status](https://travis-ci.org/skaeladmin/dfmanager.svg?branch=master)](https://travis-ci.org/skaeladmin/dfmanager)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/skaeladmin/dfmanager/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/skaeladmin/dfmanager)](https://goreportcard.com/report/github.com/skaeladmin/dfmanager)

# Dialogflow Agent Manager

## Installation
Make sure you have a working Go environment. [See Golang install instructions](http://golang.org/doc/install.html)

To install, run:
```sh
go get github.com/skaeladmin/dfmanager
```

## Use as CLI

```sh
dfmanager [global options] command [command options] [arguments...]
```

```
COMMANDS:
     export, e   exports agent from dialogflow
     import, i   imports agent to dialogflow
     restore, r  restores (replaces) agent in dialogflow
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -k value, --key value      Google Cloud Platform Auth Key [$GCP_KEY]
   -p value, --project value  Google Cloud Platform Project Name [$GCE_PROJECT]
   -f value, --file value     Input/Output file
   --help, -h                 show help
   --version, -v              print the version
```

## Use as API

```go
func main() {
	f, err := ioutil.ReadFile("someFileName")
	checkError(err)

	m, err := NewManager(f, "myDFProject")
	checkError(err)

	err = m.Export("export.zip")
	checkError(err)
}

func checkError(err error) {
	if nil != err {
		log.Fatal(err)
	}
}
```
