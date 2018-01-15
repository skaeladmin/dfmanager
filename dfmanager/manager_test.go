package dfmanager

import (
	"io/ioutil"
	"github.com/prometheus/common/log"
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
