package resources

import "fmt"

// DataImageSpecs is the struct that contains the specs of a data image
type DataImageSpecs struct {
	DiskSizeGb float64
	DeviceName string
	VolumeType string
}

// DataImageResource is the struct that contains the info of a data image resource
type DataImageResource struct {
	Identification *ResourceIdentification
	DataImageSpecs []*DataImageSpecs
}

// GetIdentification returns the identification of the resource
func (r DataImageResource) GetIdentification() *ResourceIdentification {
	return r.Identification
}

// GetAddress returns the address of the resource
func (r DataImageResource) GetAddress() string {
	return fmt.Sprintf("data.%v.%v", r.GetIdentification().ResourceType, r.GetIdentification().Name)
}

// GetKey returns the key of the resource
func (r DataImageResource) GetKey() string {
	return r.GetAddress()
}

// DataResource is the interface that contains the info of a data resource
type DataResource interface {
	GetIdentification() *ResourceIdentification
	GetAddress() string
	GetKey() string
}
