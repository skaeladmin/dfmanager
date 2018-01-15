package dfmanager

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
	"os"
	"strings"
)

//Manager in charge of all actions related to DialogFlow
type Manager struct {
	srv   *dialogflow.Service
	prj   string
	fName string
}

//NewManager parses cli context and builds DFManager instance based on provided args
func NewManager(prjKey []byte, prjName, fName string) (*Manager, error) {
	service, err := buildClient(prjKey)
	if nil != err {
		return nil, err
	}
	return &Manager{srv: service, prj: prjName, fName: fName}, nil
}

//Export downloads Dialogflow agent and saves to file
//If file name is not provided, file with GCP project name and extension .zip will be created in work directory
func (m *Manager) Export() error {
	rs, err := m.srv.Projects.Agent.Export("projects/"+m.prj, &dialogflow.ExportAgentRequest{}).Do()
	if nil != err {
		return err
	}
	if nil != rs.Error {
		return errors.New(rs.Error.Message)
	}

	var exportRS dialogflow.ExportAgentResponse
	err = json.Unmarshal(rs.Response, &exportRS)
	if nil != err {
		return err
	}

	f, err := os.Create(m.getFilename())
	if nil != err {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = w.ReadFrom(base64.NewDecoder(base64.StdEncoding, strings.NewReader(exportRS.AgentContent)))
	if nil != err {
		return err
	}

	err = w.Flush()
	if nil != err {
		return err
	}

	return nil

}

//Import reads archive and uploads it to Dialogflow
func (m *Manager) Import() error {
	f, err := os.Open(m.getFilename())
	if nil != err {
		return err
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(base64.NewEncoder(base64.StdEncoding, &buf))
	_, err = w.ReadFrom(f)
	if nil != err {
		return err
	}

	err = w.Flush()
	if nil != err {
		return err
	}
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
func (m *Manager) getFilename() string {
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

func buildClient(key []byte) (*dialogflow.Service, error) {
	jwtConfig, err := google.JWTConfigFromJSON(key, dialogflow.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("cannot build GCE compute service: %s", err)
	}
	return dialogflow.New(jwtConfig.Client(context.TODO()))
}
