package tools_gcp

import (
	"context"
	"log"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

func GetProjectId() string {
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
