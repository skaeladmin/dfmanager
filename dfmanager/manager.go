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
	srv *dialogflow.Service
	prj string
}

//NewManager parses cli context and builds DFManager instance based on provided args
func NewManager(prjKey []byte, prjName string) (*Manager, error) {
	service, err := buildClient(prjKey)
	if nil != err {
		return nil, err
	}
	return &Manager{srv: service, prj: prjName}, nil
}

//Export downloads Dialogflow agent and saves to file
//If file name is not provided, file with GCP project name and extension .zip will be created in work directory
func (m *Manager) Export(fName string) error {

	fmt.Println("Exporting agent...")
	rs, err := m.srv.Projects.Agent.Export("projects/"+m.prj, &dialogflow.GoogleCloudDialogflowV2beta1ExportAgentRequest{}).Do()
	if nil != err {
		return err
	}
	if nil != rs.Error {
		return errors.New(rs.Error.Message)
	}

	var exportRS dialogflow.GoogleCloudDialogflowV2beta1ExportAgentResponse
	err = json.Unmarshal(rs.Response, &exportRS)
	if nil != err {
		return err
	}

	f, err := os.Create(m.getFilename(fName))
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
func (m *Manager) Import(fName string) error {
	cont, err := m.readAgentContent(fName)
	if nil != err {
		return err
	}

	rq := &dialogflow.GoogleCloudDialogflowV2beta1ImportAgentRequest{}
	rq.AgentContent = cont

	fmt.Println("Importing agent from backup...")
	rs, err := m.srv.Projects.Agent.Import("projects/"+m.prj, rq).Do()
	if nil != err {
		return err
	}
	if nil != rs.Error {
		return errors.New(rs.Error.Message)
	}

	fmt.Println("Import completed successfully")
	return nil

}

//Restore reads archive and restores it in Dialogflow
func (m *Manager) Restore(fName string) error {
	cont, err := m.readAgentContent(fName)
	if nil != err {
		return err
	}

	rq := &dialogflow.GoogleCloudDialogflowV2beta1RestoreAgentRequest{}
	rq.AgentContent = cont

	fmt.Println("Restoring agent from backup...")
	rs, err := m.srv.Projects.Agent.Restore("projects/"+m.prj, rq).Do()
	if nil != err {
		return err
	}
	if nil != rs.Error {
		return errors.New(rs.Error.Message)
	}

	fmt.Println("Restore completed successfully")
	return nil

}

func (m *Manager) readAgentContent(fName string) (string, error) {
	fmt.Println("Reading agent content...")
	f, err := os.Open(m.getFilename(fName))
	if nil != err {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := bufio.NewWriter(base64.NewEncoder(base64.StdEncoding, &buf))
	_, err = w.ReadFrom(f)
	if nil != err {
		return "", err
	}

	err = w.Flush()
	if nil != err {
		return "", err
	}

	return buf.String(), nil
}

//getFilename uses provided file name or builds default one based on project name
func (m *Manager) getFilename(fName string) (name string) {
	if "" == fName {
		name = m.prj + ".zip"
	} else if !strings.HasSuffix(fName, ".zip") {
		name = fName + ".zip"
	} else {
		name = fName
	}
	return
}

func buildClient(key []byte) (*dialogflow.Service, error) {
	jwtConfig, err := google.JWTConfigFromJSON(key, dialogflow.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("cannot build GCE compute service: %s", err)
	}
	return dialogflow.New(jwtConfig.Client(context.TODO()))
}
