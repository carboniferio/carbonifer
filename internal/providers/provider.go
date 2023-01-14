package providers

import "fmt"

type Provider int

const (
	AWS = iota
	AZURE
	GCP
)

func (p Provider) String() string {
	switch p {
	case AWS:
		return "AWS"
	case AZURE:
		return "Azure"
	case GCP:
		return "GCP"
	}
	return "unknown"
}

type UnsupportedProviderError struct {
	Provider string
}

func (upe *UnsupportedProviderError) Error() string {
	return fmt.Sprintf("Unsupported Provider: %v", upe.Provider)
}
