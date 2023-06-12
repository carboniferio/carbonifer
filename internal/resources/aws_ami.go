package resources

import (
	"fmt"
)

type AmiDataResource struct {
	Identification *ResourceIdentification
	DataImageSpecs []*DataImageSpecs
	AmiId          string
}

func (r AmiDataResource) GetIdentification() *ResourceIdentification {
	return r.Identification
}

func (r AmiDataResource) GetAddress() string {
	return fmt.Sprintf("data.%v.%v", r.GetIdentification().ResourceType, r.GetIdentification().Name)
}

func (r AmiDataResource) GetKey() string {
	return r.AmiId
}
