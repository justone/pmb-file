package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

type BrokerCommand struct {
	// nothing yet
}

var brokerCommand BrokerCommand

func (x *BrokerCommand) Execute(args []string) error {
	bus := pmb.GetPMB(globalOptions.Primary)

	id := pmb.GenerateRandomID("file-broker")

	conn, err := bus.ConnectClient(id, !globalOptions.TrustKey)
	if err != nil {
		return err
	}

	return runBroker(conn)
}

func init() {
	parser.AddCommand("broker",
		"Persistent file broker, handling interfacing with S3",
		"",
		&brokerCommand)
}

func runBroker(conn *pmb.Connection) error {

	for {
		message := <-conn.In
		if message.Contents["type"].(string) == "RequestDownloadURL" {
			if latest, ok := message.Contents["latest"]; ok {
				if latest.(bool) {
					logrus.Infof("Generating S3 download url...")
					response := map[string]interface{}{
						"type":         "DownloadURLAvailable",
						"requestor":    message.Contents["id"].(string),
						"download_url": "https://s3.aws.com.url/foo.file",
					}
					conn.Out <- pmb.Message{Contents: response}
				}
			}
		} else if message.Contents["type"].(string) == "RequestUploadURL" {
			logrus.Infof("Generating S3 upload url...")
			response := map[string]interface{}{
				"type":       "UploadURLAvailable",
				"requestor":  message.Contents["id"].(string),
				"upload_url": "https://s3.aws.com.url/foo.file",
			}
			conn.Out <- pmb.Message{Contents: response}
		}
	}

	return nil
}
