package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
	"github.com/pkg/errors"
)

type GetCommand struct {
	Latest bool `short:"l" long:"latest" description:"Get the latest file that was uploaded."`
	Args   struct {
		FileName string `description:"File to download"`
	} `positional-args:"yes" required:"no"`
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

	if len(getCommand.Args.FileName) == 0 {
		return fmt.Errorf("you need to specify the filename to download, for now")
	}

	request := map[string]interface{}{
		"type":     "RequestDownloadURL",
		"filename": getCommand.Args.FileName,
	}
	conn.Out <- pmb.Message{Contents: request}

	for {
		message := <-conn.In
		if message.Contents["type"].(string) == "DownloadURLAvailable" {
			mes := UrlAvailableMessage{}
			if err := json.Unmarshal([]byte(message.Raw), &mes); err != nil {
				return errors.Wrap(err, "unable to unmarshal json")
			}

			if mes.Requestor == conn.Id && mes.Filename == getCommand.Args.FileName {
				logrus.Debugf("Going to download %s...", mes.Url)

				// build the download request
				req, err := http.NewRequest("GET", mes.Url, nil)
				if err != nil {
					return errors.Wrap(err, "unable to build presigned request")
				}

				// copy over any headers from the signing
				for key, values := range mes.Header {
					for _, val := range values {
						req.Header.Add(key, val)
					}
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return errors.Wrap(err, "unable to download file")
				}

				if resp.StatusCode != http.StatusOK {
					body, _ := ioutil.ReadAll(resp.Body)
					return fmt.Errorf("failed to download: code: %d, status: %s, body: %s", resp.StatusCode, resp.Status, string(body))
				}

				if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
					return errors.Wrap(err, "unable to save file contents")
				}

				break
			}
		}
	}

	return nil
}
