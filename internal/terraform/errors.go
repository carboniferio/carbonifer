package terraform

import "fmt"

type ProviderAuthError struct {
	ParentError error
}

func (e *ProviderAuthError) Error() string {
	return fmt.Sprintf("Missing/Invalid provider credentials, please check or set your credentials : %v", e.ParentError)
}
