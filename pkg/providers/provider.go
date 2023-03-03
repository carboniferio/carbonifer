package providers

type Provider int

const (
	AWS Provider = iota
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
	default:
		return "unknown"
	}
}
