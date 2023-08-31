package terraform

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

func getDefaultRegion() *string {
	var region interface{}
	if region == nil {
		if os.Getenv("AWS_DEFAULT_REGION") != "" {
			region = os.Getenv("AWS_DEFAULT_REGION")
		}
	}
	if region == nil {
		if os.Getenv("AWS_REGION") != "" {
			region = os.Getenv("AWS_REGION")
		}
	}

	// Check AWS Config file
	if region == nil {
		sess, err := session.NewSession()
		if err != nil {
			log.Fatalf("Error getting region from AWS config file %v", err)
		}
		if *sess.Config.Region != "" {
			region = *sess.Config.Region
		}
	}

	// Check EC2 Instance Metadata
	if region == nil {
		sess := session.Must(session.NewSession())
		svc := ec2metadata.New(sess)
		if svc.Available() {
			region, _ = svc.Region()
		}
	}
	regionPtr, ok := region.(*string)
	if ok {
		return regionPtr
	}
	regionString, ok := region.(string)
	if !ok {
		return nil
	}
	return &regionString
}
