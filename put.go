package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type PutCommand struct {
	Args struct {
		FileName string `description:"File to upload"`
	} `positional-args:"yes" required:"yes"`
}

var putCommand PutCommand

func (x *PutCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	id := pmb.GenerateRandomID("file-put")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runPut(conn)
}

func init() {
	parser.AddCommand("put",
		"Upload a file.",
		"",
		&putCommand)
}

func runPut(conn *pmb.Connection) error {

	request := map[string]interface{}{
		"type":     "RequestUploadURL",
		"filename": putCommand.Args.FileName,
	}
	conn.Out <- pmb.Message{Contents: request}

	for {
		message := <-conn.In
		if message.Contents["type"].(string) == "UploadURLAvailable" {
			if message.Contents["requestor"].(string) == conn.Id {
				logrus.Infof("Going to upload %s to %s...", putCommand.Args.FileName, message.Contents["upload_url"].(string))
				break
			}
		}
	}

	return nil
}
