package providers

import "fmt"

// ENUM(AWS, AZURE, GCP)
//
//go:generate go-enum --nocase --noprefix --marshal
type Provider int

// UnsupportedProviderError is an error that occurs when a provider is not supported
type UnsupportedProviderError struct {
	Provider string
}

func (upe *UnsupportedProviderError) Error() string {
	return fmt.Sprintf("Unsupported Provider: %v", upe.Provider)
}
