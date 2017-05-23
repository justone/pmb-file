package main

import (
	"encoding/json"
	"fmt"

	"github.com/justone/pmb/api"
	"github.com/pkg/errors"
)

type ListCommand struct {
	Count int64 `short:"c" long:"count" description:"How many files to list" default:"10"`
}

var listCommand ListCommand

func (x *ListCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	id := pmb.GenerateRandomID("file-list")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runList(conn)
}

func init() {
	parser.AddCommand("list",
		"Retrieve a file.",
		"",
		&listCommand)
}

func runList(conn *pmb.Connection) error {

	request := map[string]interface{}{
		"type":  "RequestFileList",
		"count": listCommand.Count,
	}
	conn.Out <- pmb.Message{Contents: request}

	for {
		message := <-conn.In
		if message.Contents["type"].(string) == "FileListing" {
			mes := FileListingMessage{}
			if err := json.Unmarshal([]byte(message.Raw), &mes); err != nil {
				return errors.Wrap(err, "unable to unmarshal json")
			}

			// if the listing is for us
			if mes.Requestor == conn.Id {
				// print out files
				for idx, file := range mes.Files {
					fmt.Printf("%d: %s (size: %d, uploaded: %v)\n", idx, file.Name, file.Size, file.Modified)
				}

				// and we're done looking for messages
				break
			}
		}
	}

	return nil
}
