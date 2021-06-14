// Copyright 2021 Rubrik, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package gcp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

type project struct {
	id      string
	name    string
	number  int64
	orgName string
	creds   *google.Credentials
}

// ProjectFunc returns a project initialized from the values passed to the
// function creating the ProjectFunc.
type ProjectFunc func(ctx context.Context) (project, error)

// readCredentials reads the credentials from the specified key file.
func readCredentials(ctx context.Context, keyFile string) (*google.Credentials, error) {
	if strings.HasPrefix(keyFile, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		keyFile = filepath.Join(home, strings.TrimPrefix(keyFile, "~/"))
	}

	buf, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	creds, err := google.CredentialsFromJSON(ctx, buf, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}

	return creds, nil
}

// gcpProject returns a project initialized values from the credentials and the
// cloud.
func gcpProject(ctx context.Context, creds *google.Credentials, id string) (project, error) {
	if id == "" {
		return project{}, errors.New("polaris: project id cannot be empty")
	}

	client, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return project{}, err
	}

	// Lookup project.
	proj, err := client.Projects.Get(id).Do()
	if err != nil {
		return project{}, err
	}

	// Lookup parent organization.
	orgName := proj.Parent.Id
	if proj.Parent.Type == "organization" {
		org, err := client.Organizations.Get("organizations/" + proj.Parent.Id).Do()
		if err != nil {
			return project{}, err
		}

		if org.DisplayName != "" {
			orgName = org.DisplayName
		}
	}

	project := project{
		id:      id,
		name:    proj.Name,
		number:  proj.ProjectNumber,
		orgName: orgName,
		creds:   creds,
	}

	return project, nil
}

// Credentials returns a ProjectFunc that initializes the project with values
// from the specified credentials and the cloud using the credentials project
// id.
func Credentials(credentials *google.Credentials) ProjectFunc {
	return func(ctx context.Context) (project, error) {
		return gcpProject(ctx, credentials, credentials.ProjectID)
	}
}

// Default returns a ProjectFunc that initializes the project with values from
// the default credentials and the cloud using the default credentials project
// id.
func Default() ProjectFunc {
	return func(ctx context.Context) (project, error) {
		creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return project{}, err
		}

		return gcpProject(ctx, creds, creds.ProjectID)
	}
}

// KeyFile returns a ProjectFunc that initializes the project with values from
// the specified key file and the cloud using the key file project id.
func KeyFile(keyFile string) ProjectFunc {
	return func(ctx context.Context) (project, error) {
		creds, err := readCredentials(ctx, keyFile)
		if err != nil {
			return project{}, err
		}

		return gcpProject(ctx, creds, creds.ProjectID)
	}
}

// KeyFileAndProject returns a ProjectFunc that initializes the project with
// values from the specified key file and the cloud using the given project id.
func KeyFileAndProject(keyFile, projectID string) ProjectFunc {
	return func(ctx context.Context) (project, error) {
		creds, err := readCredentials(ctx, keyFile)
		if err != nil {
			return project{}, err
		}

		return gcpProject(ctx, creds, projectID)
	}
}

// Project returns a ProjectFunc that initializes the project with the
// specified values.
func Project(id, name string, number int64, orgName string) ProjectFunc {
	return func(ctx context.Context) (project, error) {
		return project{id: id, number: number, name: name, orgName: orgName}, nil
	}
}
