package main

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/justone/pmb/api"
	"github.com/pkg/errors"
)

type BrokerCommand struct {
	Bucket string `long:"s3-bucket" description:"S3 bucket to use"`
	Region string `env:"AWS_DEFAULT_REGION" long:"s3-region" description:"S3 region"`
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

	sess := session.Must(session.NewSession())

	var region string
	if len(brokerCommand.Region) == 0 {
		var err error
		region, err = s3manager.GetBucketRegion(aws.BackgroundContext(), sess, brokerCommand.Bucket, endpoints.UsWest2RegionID)
		if err != nil {
			return errors.Wrap(err, "unable to determine bucket region")
		}
	}
	s3svc := s3.New(sess, &aws.Config{
		Region: aws.String(region),
	})

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
			filename := message.Contents["filename"].(string)

			putObjReq, _ := s3svc.PutObjectRequest(&s3.PutObjectInput{
				Bucket:        aws.String(brokerCommand.Bucket),
				Key:           aws.String(filename),
				ContentLength: aws.Int64(0),
			})
			url, headers, err := putObjReq.PresignRequest(15 * time.Minute)

			if err != nil {
				logrus.Warnf("error presigning: %v", err)
			}

			logrus.Infof("Generating S3 upload url...")
			response := map[string]interface{}{
				"type":       "UploadURLAvailable",
				"requestor":  message.Contents["id"].(string),
				"filename":   filename,
				"upload_url": url,
				"headers":    headers,
			}
			conn.Out <- pmb.Message{Contents: response}
		}
	}

	return nil
}
