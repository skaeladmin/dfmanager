package main

import (
	"fmt"
	"github.com/skaeladmin/dfmanager/dfmanager"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"os"
)

var version = "master"

func main() {
	app := cli.NewApp()
	app.Name = "dfmanager"
	app.Usage = "Dialogflow Agent Manager"
	app.Version = version

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
		exportCommand,
		importCommand,
		restoreCommand,
	}

	err := app.Run(os.Args)
	if nil != err {
		log.Fatal(err)
	}
}

//NewCliDFManager parses cli context and builds DFManager instance based on provided args
func NewCliDFManager(c *cli.Context) (*dfmanager.Manager, error) {
	key, err := getKey(c)
	if nil != err {
		return nil, err
	}

	project, err := getArg(c, "project", true)
	if nil != err {
		return nil, err
	}

	return dfmanager.NewManager(key, project)
}

//getFile parses filename from incoming parameters
func getFile(c *cli.Context) (string, error) {
	f, err := getArg(c, "file", false)
	if nil != err {
		return "", err
	}
	return f, nil
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

var (
	exportCommand = cli.Command{
		Name:    "export",
		Aliases: []string{"e"},
		Usage:   "exports agent from dialogflow",
		Action: func(c *cli.Context) error {
			manager, err := NewCliDFManager(c)
			if nil != err {
				return cli.NewExitError(err, 1)
			}
			f, err := getFile(c)
			if nil != err {
				return cli.NewExitError(err, 1)
			}

			err = manager.ExportToFile(f)
			if nil != err {
				return cli.NewExitError(err, 1)
			}

			return nil
		},
	}

	importCommand = cli.Command{
		Name:    "import",
		Aliases: []string{"i"},
		Usage:   "imports agent to dialogflow",
		Action: func(c *cli.Context) error {
			manager, err := NewCliDFManager(c)
			if nil != err {
				return cli.NewExitError(err, 1)
			}
			f, err := getFile(c)
			if nil != err {
				return cli.NewExitError(err, 1)
			}

			err = manager.Import(f)
			if nil != err {
				return cli.NewExitError(err, 1)
			}

			return nil
		},
	}

	restoreCommand = cli.Command{
		Name:    "restore",
		Aliases: []string{"r"},
		Usage:   "restores (replaces) agent in dialogflow",
		Action: func(c *cli.Context) error {
			manager, err := NewCliDFManager(c)
			if nil != err {
				return cli.NewExitError(err, 1)
			}
			f, err := getFile(c)
			if nil != err {
				return cli.NewExitError(err, 1)
			}

			err = manager.Restore(f)
			if nil != err {
				return cli.NewExitError(err, 1)
			}

			return nil
		},
	}
)
