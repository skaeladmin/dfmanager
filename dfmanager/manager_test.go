package dfmanager

import (
	"io/ioutil"
	"testing"

	"github.com/prometheus/common/log"
)

func ExampleManager() {
	f, err := ioutil.ReadFile("someFileName")
	checkError(err)

	m, err := NewManager(f, "myDFProject")
	checkError(err)

	err = m.ExportToFile("export.zip")
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func TestManager_getFilename(t *testing.T) {
	type fields struct {
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
		ttt := tt
		t.Run(ttt.name, func(t *testing.T) {
			m := &Manager{
				prj: ttt.fields.prj,
			}
			if got := m.getFilename(ttt.fields.fName); got != ttt.want {
				t.Errorf("Manager.getFilename() = %v, want %v", got, ttt.want)
			}
		})
	}
}
