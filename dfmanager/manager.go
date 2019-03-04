package dfmanager

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	dfproto "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"

	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	"github.com/prometheus/common/log"
)

//Manager in charge of all actions related to DialogFlow
type Manager struct {
	ac  *dialogflow.AgentsClient
	ec  *dialogflow.EntityTypesClient
	prj string
}

//NewManager parses cli context and builds DFManager instance based on provided args
func NewManager(prjKey []byte, prjName string) (*Manager, error) {
	ac, ec, err := buildClient(prjKey)
	if err != nil {
		return nil, err
	}
	return &Manager{ac: ac, ec: ec, prj: prjName}, nil
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
	_, err = w.ReadFrom(bytes.NewReader(content))
	if err != nil {
		return err
	}

	err = w.Flush()
	if err != nil {
		return err
	}

	return nil

}

//Export downloads Dialogflow agent and returns it as zip archive byte array
func (m *Manager) Export() ([]byte, error) {

	fmt.Println("Exporting agent...")
	expOperation, err := m.ac.ExportAgent(context.Background(), &dfproto.ExportAgentRequest{
		Parent: "projects/" + m.prj,
	})
	if err != nil {
		return nil, err
	}
	rs, err := expOperation.Wait(context.Background())
	if err != nil {
		return nil, err
	}

	return rs.GetAgentContent(), nil
}

//ImportFile reads archive and uploads it to Dialogflow
func (m *Manager) ImportFile(fName string) error {
	cont, err := m.readAgentContent(fName)
	if err != nil {
		return err
	}
	return m.Import(cont)
}

//Import expects content to be zip file byte array agent content
func (m *Manager) Import(content []byte) error {

	fmt.Println("Importing agent from backup...")
	rs, err := m.ac.ImportAgent(context.Background(), &dfproto.ImportAgentRequest{
		Parent: "projects/" + m.prj,
		Agent: &dfproto.ImportAgentRequest_AgentContent{
			AgentContent: content,
		},
	})
	if err != nil {
		return err
	}

	if err := rs.Wait(context.Background()); err != nil {
		return err
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
func (m *Manager) ListEntityTypes() ([]*dfproto.EntityType, error) {
	fmt.Println("List entity types...")
	rs := m.ec.ListEntityTypes(context.Background(), &dfproto.ListEntityTypesRequest{
		Parent: "projects/" + m.prj + "/agent",
	})

	var all []*dfproto.EntityType
	for next, err := rs.Next(); err != nil; {
		all = append(all, next)
	}

	return all, nil
}

//BatchUpdateEntities updates entities for one given group in batch manner
func (m *Manager) BatchUpdateEntities(name string, entities []*dfproto.EntityType_Entity) error {
	rs, err := m.ec.BatchUpdateEntities(context.Background(),
		&dfproto.BatchUpdateEntitiesRequest{
			Parent:   "projects/" + m.prj + "/agent",
			Entities: entities,
		})
	if err != nil {
		return err
	}
	err = rs.Wait(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("Entities udpated successfully")
	return nil
}

//Restore reads content (agent zip archive) and restores it in Dialogflow
func (m *Manager) Restore(content []byte) error {

	fmt.Println("Restoring agent from backup...")
	rs, err := m.ac.RestoreAgent(context.Background(), &dfproto.RestoreAgentRequest{
		Parent: "projects/" + m.prj,
		Agent: &dfproto.RestoreAgentRequest_AgentContent{
			AgentContent: content,
		},
	})
	if err != nil {
		return err
	}
	if err := rs.Wait(context.Background()); err != nil {
		return err
	}

	fmt.Println("Restore completed successfully")
	return nil

}

func (m *Manager) readAgentContent(fName string) ([]byte, error) {
	fmt.Println("Reading agent content...")
	f, err := os.Open(m.getFilename(fName))
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Error(closeErr)
		}
	}()

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	_, err = w.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	err = w.Flush()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

//getFilename uses provided file name or builds default one based on project name
func (m *Manager) getFilename(fName string) (name string) {
	fmt.Println(fName)
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

func buildClient(key []byte) (*dialogflow.AgentsClient, *dialogflow.EntityTypesClient, error) {

	jwtConfig, err := google.JWTConfigFromJSON(key, compute.CloudPlatformScope)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot build GCE compute service: %s", err)
	}

	auth := option.WithTokenSource(jwtConfig.TokenSource(context.Background()))
	ac, err := dialogflow.NewAgentsClient(context.Background(), auth)
	if err != nil {
		return nil, nil, err
	}
	ec, err := dialogflow.NewEntityTypesClient(context.Background(), auth)
	if err != nil {
		return nil, nil, err
	}
	return ac, ec, err
}
