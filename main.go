package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dialogflow/v2beta1"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	app := cli.NewApp()
	app.Name = "dfmanager"
	app.Usage = "Dialogflow Agent Manager"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "k, key",
			Usage:  "Google Cloud Platform Auth Key",
			EnvVar: "GCP_KEY",
		},
		cli.StringFlag{
			Name:   "p, project",
			Usage:  "Google Cloud Platform Project Name",
			EnvVar: "GCE_PROJECT",
		},

		cli.StringFlag{
			Name:  "f, file",
			Usage: "Input/Output file",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "export",
			Aliases: []string{"e"},
			Usage:   "exports agent from dialogflow",
			Action: func(c *cli.Context) error {
				manager, err := NewCliDFManager(c)
				if nil != err {
					return cli.NewExitError(err, 1)
				}

				err = manager.Export()
				if nil != err {
					return cli.NewExitError(err, 1)
				}

				return nil
			},
		},
		{
			Name:    "import",
			Aliases: []string{"i"},
			Usage:   "imports agent to dialogflow",
			Action: func(c *cli.Context) error {
				manager, err := NewCliDFManager(c)
				if nil != err {
					return cli.NewExitError(err, 1)
				}

				err = manager.Import()
				if nil != err {
					return cli.NewExitError(err, 1)
				}

				return nil
			},
		},
	}

	app.Run(os.Args)
}

//DFManager in charge of all actions related to DialogFlow
type DFManager struct {
	srv   *dialogflow.Service
	prj   string
	fName string
}

//NewDFManager parses cli context and builds DFManager instance based on provided args
func NewDFManager(prjKey []byte, prjName, fName string) (*DFManager, error) {
	service, err := buildClient(prjKey)
	if nil != err {
		return nil, err
	}
	return &DFManager{srv: service, prj: prjName, fName: fName}, nil
}

//NewCliDFManager parses cli context and builds DFManager instance based on provided args
func NewCliDFManager(c *cli.Context) (*DFManager, error) {
	key, err := getKey(c)
	if nil != err {
		return nil, err
	}

	project, err := getArg(c, "project", true)
	if nil != err {
		return nil, err
	}

	f, err := getArg(c, "file", false)
	if nil != err {
		return nil, err
	}

	return NewDFManager(key, project, f)
}

//Export downloads Dialogflow agent and saves to file
//If file name is not provided, file with GCP project name and extension .zip will be created in work directory
func (m *DFManager) Export() error {
	rs, err := m.srv.Projects.Agent.Export("projects/"+m.prj, &dialogflow.ExportAgentRequest{}).Do()
	if nil != err {
		return err
	}
	if nil != rs.Error {
		return errors.New(rs.Error.Message)
	}

	var exportRS dialogflow.ExportAgentResponse
	json.Unmarshal(rs.Response, &exportRS)

	f, err := os.Create(m.getFilename())
	if nil != err {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.ReadFrom(base64.NewDecoder(base64.StdEncoding, strings.NewReader(exportRS.AgentContent)))
	w.Flush()

	return nil

}

//Import reads archive and uploads it to Dialogflow
func (m *DFManager) Import() error {
	f, err := os.Open(m.getFilename())
	if nil != err {
		return err
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(base64.NewEncoder(base64.StdEncoding, &buf))
	w.ReadFrom(f)
	w.Flush()
	defer f.Close()

	rq := &dialogflow.ImportAgentRequest{}
	rq.AgentContent = buf.String()

	rs, err := m.srv.Projects.Agent.Import("projects/"+m.prj, rq).Do()
	if nil != err {
		return err
	}
	if nil != rs.Error {
		return errors.New(rs.Error.Message)
	}

	return nil

}

//getFilename uses provided file name or builds default one based on project name
func (m *DFManager) getFilename() string {
	var fname string
	if "" == m.fName {
		fname = m.prj + ".zip"
	} else if !strings.HasSuffix(m.fName, ".zip") {
		fname = m.fName + ".zip"
	} else {
		fname = m.fName
	}
	return fname
}

//getArg grabs arg value from the context
//if argument is required, error in case of empty value will be sent
func getArg(c *cli.Context, name string, required bool) (string, error) {
	var val string
	if val = c.GlobalString(name); "" == val && required {
		return "", fmt.Errorf("argument %s is missed", name)
	}
	return val, nil
}

func getKey(c *cli.Context) ([]byte, error) {
	key, err := getArg(c, "key", true)
	if nil != err {
		return nil, err
	}
	return ioutil.ReadFile(key)
}

func buildClient(key []byte) (*dialogflow.Service, error) {
	jwtConfig, err := google.JWTConfigFromJSON(key, dialogflow.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("cannot build GCE compute service: %s", err)
	}
	return dialogflow.New(jwtConfig.Client(context.TODO()))
}
