package resources

import "fmt"

type DataImageSpecs struct {
	DiskSizeGb float64
}

type DataImageResource struct {
	Identification *ResourceIdentification
	DataImageSpecs *DataImageSpecs
}

func (r DataImageResource) GetIdentification() *ResourceIdentification {
	return r.Identification
}

func (r DataImageResource) GetAddress() string {
	return fmt.Sprintf("data.%v.%v", r.GetIdentification().ResourceType, r.GetIdentification().Name)
}

func (r DataImageResource) GetKey() string {
	return r.GetAddress()
}

type DataResource interface {
	GetIdentification() *ResourceIdentification
	GetAddress() string
	GetKey() string
}
