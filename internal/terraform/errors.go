package terraform

import "fmt"

// ProviderAuthError is the struct that contains the error of a provider auth error
type ProviderAuthError struct {
	ParentError error
}

// Error returns the error of a provider auth error
func (e *ProviderAuthError) Error() string {
	return fmt.Sprintf("Missing/Invalid provider credentials, please check or set your credentials : %v", e.ParentError)
}
