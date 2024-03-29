// Code generated by go-enum DO NOT EDIT.
// Version:
// Revision:
// Build Date:
// Built By:

package providers

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// AWS is a Provider of type AWS.
	AWS Provider = iota
	// AZURE is a Provider of type AZURE.
	AZURE
	// GCP is a Provider of type GCP.
	GCP
)

var ErrInvalidProvider = errors.New("not a valid Provider")

const _ProviderName = "AWSAZUREGCP"

var _ProviderMap = map[Provider]string{
	AWS:   _ProviderName[0:3],
	AZURE: _ProviderName[3:8],
	GCP:   _ProviderName[8:11],
}

// String implements the Stringer interface.
func (x Provider) String() string {
	if str, ok := _ProviderMap[x]; ok {
		return str
	}
	return fmt.Sprintf("Provider(%d)", x)
}

var _ProviderValue = map[string]Provider{
	_ProviderName[0:3]:                   AWS,
	strings.ToLower(_ProviderName[0:3]):  AWS,
	_ProviderName[3:8]:                   AZURE,
	strings.ToLower(_ProviderName[3:8]):  AZURE,
	_ProviderName[8:11]:                  GCP,
	strings.ToLower(_ProviderName[8:11]): GCP,
}

// ParseProvider attempts to convert a string to a Provider.
func ParseProvider(name string) (Provider, error) {
	if x, ok := _ProviderValue[name]; ok {
		return x, nil
	}
	// Case insensitive parse, do a separate lookup to prevent unnecessary cost of lowercasing a string if we don't need to.
	if x, ok := _ProviderValue[strings.ToLower(name)]; ok {
		return x, nil
	}
	return Provider(0), fmt.Errorf("%s is %w", name, ErrInvalidProvider)
}

// MarshalText implements the text marshaller method.
func (x Provider) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *Provider) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseProvider(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
