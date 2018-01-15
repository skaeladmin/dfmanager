package dfmanager

import (
	"io/ioutil"
	"testing"

	"github.com/prometheus/common/log"
	"google.golang.org/api/dialogflow/v2beta1"
)

func ExampleDfManager() {
	f, err := ioutil.ReadFile("someFileName")
	checkError(err)

	m, err := NewManager(f, "myDFProject", "export.zip")
	checkError(err)

	err = m.Export()
	checkError(err)
}

func checkError(err error) {
	if nil != err {
		log.Fatal(err)
	}
}

func TestManager_getFilename(t *testing.T) {
	type fields struct {
		srv   *dialogflow.Service
		prj   string
		fName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "filename provided with extension",
			fields: fields{fName: "file", prj: "project"},
			want:   "file.zip",
		},
		{
			name:   "filename provided, no extension",
			fields: fields{fName: "file", prj: "project"},
			want:   "file.zip",
		},
		{
			name:   "filename not provided",
			fields: fields{prj: "project"},
			want:   "project.zip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				srv:   tt.fields.srv,
				prj:   tt.fields.prj,
				fName: tt.fields.fName,
			}
			if got := m.getFilename(); got != tt.want {
				t.Errorf("Manager.getFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
