package dfmanager

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/prometheus/common/log"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dialogflow/v2"
)

//Manager in charge of all actions related to DialogFlow
type Manager struct {
	srv *dialogflow.Service
	prj string
}

//NewManager parses cli context and builds DFManager instance based on provided args
func NewManager(prjKey []byte, prjName string) (*Manager, error) {
	service, err := buildClient(prjKey)
	if err != nil {
		return nil, err
	}
	return &Manager{srv: service, prj: prjName}, nil
}

//ExportToFile downloads Dialogflow agent and saves to file
//If file name is not provided, file with GCP project name and extension .zip will be created in work directory
func (m *Manager) ExportToFile(fName string) error {

	content, err := m.Export()
	if err != nil {
		return err
	}

	f, err := os.Create(m.getFilename(fName))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Fatal(closeErr)
		}
	}()

	w := bufio.NewWriter(f)
	_, err = w.ReadFrom(base64.NewDecoder(base64.StdEncoding, strings.NewReader(content)))
	if err != nil {
		return err
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	return nil

}

//Export downloads Dialogflow agent and returns it as BASE64 encoded zip archive string
func (m *Manager) Export() (string, error) {

	fmt.Println("Exporting agent...")
	rs, err := m.srv.Projects.Agent.Export("projects/"+m.prj, &dialogflow.GoogleCloudDialogflowV2ExportAgentRequest{}).Do()
	if err != nil {
		return "", err
	}
	if rs.Error != nil {
		return "", errors.New(rs.Error.Message)
	}

	var exportRS dialogflow.GoogleCloudDialogflowV2beta1ExportAgentResponse
	err = json.Unmarshal(rs.Response, &exportRS)
	if err != nil {
		return "", err
	}

	return exportRS.AgentContent, nil
}

//ImportFile reads archive and uploads it to Dialogflow
func (m *Manager) ImportFile(fName string) error {
	cont, err := m.readAgentContent(fName)
	if err != nil {
		return err
	}
	return m.Import(cont)
}

//Import expects content to be BASE64 encoded zip agent content
func (m *Manager) Import(content string) error {
	rq := &dialogflow.GoogleCloudDialogflowV2ImportAgentRequest{}
	rq.AgentContent = content

	fmt.Println("Importing agent from backup...")
	rs, err := m.srv.Projects.Agent.Import("projects/"+m.prj, rq).Do()
	if err != nil {
		return err
	}
	if rs.Error != nil {
		return errors.New(rs.Error.Message)
	}

	fmt.Println("Import completed successfully")
	return nil

}

//RestoreFile reads archive and restores it in Dialogflow
func (m *Manager) RestoreFile(fName string) error {
	cont, err := m.readAgentContent(fName)
	if err != nil {
		return err
	}
	return m.Restore(cont)

}

//ListEntityTypes reads and returns list of all entities that belong to the project
func (m *Manager) ListEntityTypes() ([]*dialogflow.GoogleCloudDialogflowV2EntityType, error) {
	fmt.Println("List entity types...")
	rs, err := m.srv.Projects.Agent.EntityTypes.List("projects/" + m.prj + "/agent").Do()
	if err != nil {
		return nil, err
	}
	return rs.EntityTypes, nil
}

//BatchUpdateEntities updates entities for one given group in batch manner
func (m *Manager) BatchUpdateEntities(name string, entities []*dialogflow.GoogleCloudDialogflowV2EntityTypeEntity) error {
	rs, err := m.srv.Projects.Agent.EntityTypes.BatchUpdate(
		"projects/"+m.prj+"/agent",
		&dialogflow.GoogleCloudDialogflowV2BatchUpdateEntityTypesRequest{
			EntityTypeBatchInline: &dialogflow.GoogleCloudDialogflowV2EntityTypeBatch{
				EntityTypes: []*dialogflow.GoogleCloudDialogflowV2EntityType{{
					Entities: entities,
					Name:     name,
				}}},
		}).Do()
	if err != nil {
		return err
	}
	if rs.Error != nil {
		return errors.New(rs.Error.Message)
	}

	fmt.Println("Entities udpated successfully")
	return nil
}

//Restore reads content (BASE64 encoded agent zip archive) and restores it in Dialogflow
func (m *Manager) Restore(content string) error {

	rq := &dialogflow.GoogleCloudDialogflowV2RestoreAgentRequest{}
	rq.AgentContent = content

	fmt.Println("Restoring agent from backup...")
	rs, err := m.srv.Projects.Agent.Restore("projects/"+m.prj, rq).Do()
	if err != nil {
		return err
	}
	if rs.Error != nil {
		return errors.New(rs.Error.Message)
	}

	fmt.Println("Restore completed successfully")
	return nil

}

func (m *Manager) readAgentContent(fName string) (string, error) {
	fmt.Println("Reading agent content...")
	f, err := os.Open(m.getFilename(fName))
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Error(closeErr)
		}
	}()

	var buf bytes.Buffer
	w := bufio.NewWriter(base64.NewEncoder(base64.StdEncoding, &buf))
	_, err = w.ReadFrom(f)
	if err != nil {
		return "", err
	}

	err = w.Flush()
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

//getFilename uses provided file name or builds default one based on project name
func (m *Manager) getFilename(fName string) (name string) {

	switch {
	case fName == "":
		name = m.prj + ".zip"
	case !strings.HasSuffix(fName, ".zip"):
		name = fName + ".zip"
	default:
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
