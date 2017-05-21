package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
	"github.com/pkg/errors"
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

	file, err := os.Open(putCommand.Args.FileName)
	if err != nil {
		return errors.Wrap(err, "unable to open file")
	}

	stat, err := file.Stat()
	if err != nil {
		return errors.Wrapf(err, "failed to stat file %s")
	}

	size := stat.Size()

	request := map[string]interface{}{
		"type":     "RequestUploadURL",
		"filename": putCommand.Args.FileName,
	}
	conn.Out <- pmb.Message{Contents: request}

	for {
		message := <-conn.In
		if message.Contents["type"].(string) == "UploadURLAvailable" {
			mes := UploadAvailableMessage{}
			if err := json.Unmarshal([]byte(message.Raw), &mes); err != nil {
				return errors.Wrap(err, "unable to unmarshal json")
			}

			// if the upload url message we received was for us and the right filename...
			if mes.Requestor == conn.Id && mes.Filename == putCommand.Args.FileName {
				logrus.Infof("Going to upload %s to %s...", putCommand.Args.FileName, mes.Url)

				// build the upload request
				req, err := http.NewRequest("PUT", mes.Url, nil)
				if err != nil {
					return errors.Wrap(err, "unable to build presigned request")
				}

				// copy over any headers from the signing
				for key, values := range mes.Header {
					for _, val := range values {
						req.Header.Add(key, val)
					}
				}

				// set the content length and contnt
				req.ContentLength = size
				req.Body = file

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return errors.Wrap(err, "unable to upload file")
				}

				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					body, _ := ioutil.ReadAll(resp.Body)
					return fmt.Errorf("failed to upload: code: %d, status: %s, body: %s", resp.StatusCode, resp.Status, string(body))
				}

				// upload is done, and we're done looking for messages
				break
			}
		}
	}

	return nil
}

type UploadAvailableMessage struct {
	Type      string      `json:"type"`
	Requestor string      `json:"requestor"`
	Filename  string      `json:"filename"`
	Url       string      `json:"upload_url"`
	Header    http.Header `json:"headers"`
}
