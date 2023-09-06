package resources

type EbsDataResource struct {
	Identification *ResourceIdentification
	DataImageSpecs []*DataImageSpecs
	AwsId          string
}

func (r EbsDataResource) GetIdentification() *ResourceIdentification {
	return r.Identification
}

func (r EbsDataResource) GetAddress() string {
	return r.Identification.Address
}

func (r EbsDataResource) GetKey() string {
	return r.AwsId
}
