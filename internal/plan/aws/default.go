package aws

import (
	"os"

	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	"github.com/carboniferio/carbonifer/internal/utils"
	tfjson "github.com/hashicorp/terraform-json"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

func GetDefaults(awsConfig *tfjson.ProviderConfig, tfPlan *tfjson.Plan, terraformRefs *tfrefs.References) {
	log.Debugf("Reading provider config %v", awsConfig.Name)

	region := getDefaultRegion(awsConfig, tfPlan)
	if region != nil {
		terraformRefs.ProviderConfigs["region"] = *region
	}
}

func getDefaultRegion(awsConfig *tfjson.ProviderConfig, tfPlan *tfjson.Plan) *string {
	var region interface{}
	regionExpr := awsConfig.Expressions["region"]
	if regionExpr != nil {
		var err error
		region, err = utils.GetValueOfExpression(regionExpr, tfPlan)
		if err != nil {
			log.Fatalf("Error getting region from provider config %v", err)
		}
	}
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
	regionString := region.(string)
	return &regionString
}
