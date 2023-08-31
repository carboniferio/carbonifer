package toolsgcp

import (
	"context"
	"log"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

// GetProjectID returns the project ID of the current GCP project
func GetProjectID() string {
	ctx := context.Background()

	// Get the default client using the default credentials
	client, err := google.DefaultClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	crmService, err := cloudresourcemanager.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal(err)
	}

	// Set the project ID
	projects, err := crmService.Projects.List().Do()
	if err != nil {
		log.Fatal(err)
	}
	return projects.Projects[0].Name
}
