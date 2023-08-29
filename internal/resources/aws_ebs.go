package resources

import (
	"fmt"
)

type EbsDataResource struct {
	Identification *ResourceIdentification
	DataImageSpecs []*DataImageSpecs
	AwsId          string
}

func (r EbsDataResource) GetIdentification() *ResourceIdentification {
	return r.Identification
}

func (r EbsDataResource) GetAddress() string {
	return fmt.Sprintf("data.%v.%v", r.GetIdentification().ResourceType, r.GetIdentification().Name)
}

func (r EbsDataResource) GetKey() string {
	return r.AwsId
}
