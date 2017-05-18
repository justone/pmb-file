package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type GetCommand struct {
	Latest bool `short:"l" long:"latest" description:"Get the latest file that was uploaded."`
}

var getCommand GetCommand

func (x *GetCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	id := pmb.GenerateRandomID("file-get")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runGet(conn)
}

func init() {
	parser.AddCommand("get",
		"Retrieve a file.",
		"",
		&getCommand)
}

func runGet(conn *pmb.Connection) error {

	request := map[string]interface{}{
		"type":   "RequestDownloadURL",
		"latest": true,
	}
	conn.Out <- pmb.Message{Contents: request}

	for {
		message := <-conn.In
		if message.Contents["type"].(string) == "DownloadURLAvailable" {
			if message.Contents["requestor"].(string) == conn.Id {
				logrus.Infof("Going to download %s...", message.Contents["download_url"].(string))
				break
			}
		}
	}

	return nil
}
