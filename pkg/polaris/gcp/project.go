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
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2/google"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

const scopes = "https://www.googleapis.com/auth/cloud-platform"

type project struct {
	NativeID string
	name     string
	number   int64
	orgName  string
	creds    *google.Credentials
}

// ProjectFunc returns a project initialized from the values passed to the
// function creating the ProjectFunc.
type ProjectFunc func(ctx context.Context) (project, error)

// Credentials returns a ProjectFunc that initializes the project with values
// from the specified credentials and the cloud using the credentials project
// ID.
// Note, the entity identified by the credentials passed into this function
// requires the resourcemanager.projects.get permission to be able to read the
// project name and project number.
func Credentials(credentials *google.Credentials) ProjectFunc {
	return func(ctx context.Context) (project, error) {
		return gcpProject(ctx, credentials, credentials.ProjectID)
	}
}

// Default returns a ProjectFunc that initializes the project with values from
// the default credentials and the cloud using the default credentials project
// ID.
// Note, the entity identified by the default credentials requires the
// resourcemanager.projects.get permission to be able to read the project name
// and project number.
func Default() ProjectFunc {
	return func(ctx context.Context) (project, error) {
		creds, err := google.FindDefaultCredentials(ctx, scopes)
		if err != nil {
			return project{}, fmt.Errorf("failed to find the default GCP credentials: %s", err)
		}

		return gcpProject(ctx, creds, creds.ProjectID)
	}
}

// Key returns a ProjectFunc that initializes the project with the credentials
// and project ID from the key. The key can be either a base64 encoded key or
// the file path to a key.
func Key(key string) ProjectFunc {
	return func(ctx context.Context) (project, error) {
		keyData, textErr := base64.StdEncoding.DecodeString(key)
		if textErr == nil {
			var creds *google.Credentials
			creds, textErr = google.CredentialsFromJSON(ctx, keyData, scopes)
			if textErr == nil {
				return project{NativeID: creds.ProjectID, creds: creds}, nil
			}
		}

		creds, fileErr := readCredentials(ctx, key)
		if fileErr != nil {
			return project{}, fmt.Errorf("credentials are neither a valid base64 encoded key (err: %s) or a path to a file containing a valid key (err: %s)", textErr, fileErr)
		}

		return project{NativeID: creds.ProjectID, creds: creds}, nil
	}
}

// KeyWithProject returns a ProjectFunc that initializes the project with the
// credentials from the key and the specified project ID. The key can be either
// a base64 encoded key or the file path to a key.
func KeyWithProject(key string, projectID string) ProjectFunc {
	return func(ctx context.Context) (project, error) {
		keyData, textErr := base64.StdEncoding.DecodeString(key)
		if textErr == nil {
			var creds *google.Credentials
			creds, textErr = google.CredentialsFromJSON(ctx, keyData, scopes)
			if textErr == nil {
				return project{NativeID: creds.ProjectID, creds: creds}, nil
			}
		}

		creds, fileErr := readCredentials(ctx, key)
		if fileErr != nil {
			return project{}, fmt.Errorf("credentials are neither a valid base64 encoded key (err: %s) or a path to a file containing a valid key (err: %s)", textErr, fileErr)
		}

		return project{NativeID: creds.ProjectID, creds: creds}, nil
	}
}

// KeyWithProjectAndNumber returns a ProjectFunc that initializes the project
// with the credentials from the key and the specified project ID and number.
// The key can be either a base64 encoded key or the file path to a key.
func KeyWithProjectAndNumber(key string, projectID string, projectNumber int64) ProjectFunc {
	return func(ctx context.Context) (project, error) {
		keyData, textErr := base64.StdEncoding.DecodeString(key)
		if textErr == nil {
			var creds *google.Credentials
			creds, textErr = google.CredentialsFromJSON(ctx, keyData, scopes)
			if textErr == nil {
				return project{NativeID: projectID, creds: creds, number: projectNumber}, nil
			}
		}

		creds, fileErr := readCredentials(ctx, key)
		if fileErr != nil {
			return project{}, fmt.Errorf("credentials are neither a valid base64 encoded key (err: %s) or a path to a file containing a valid key (err: %s)", textErr, fileErr)
		}

		return project{NativeID: projectID, creds: creds, number: projectNumber}, nil
	}
}

// KeyFile returns a ProjectFunc that initializes the project with values from
// the specified key file and the cloud using the key file's project ID.
// Note, the entity identified by the key file passed into this function
// requires the resourcemanager.projects.get permission to be able to read the
// project name and project number.
func KeyFile(keyFile string) ProjectFunc {
	return func(ctx context.Context) (project, error) {
		creds, err := readCredentials(ctx, keyFile)
		if err != nil {
			return project{}, fmt.Errorf("failed to read credentials: %s", err)
		}

		return gcpProject(ctx, creds, creds.ProjectID)
	}
}

// KeyFileWithProject returns a ProjectFunc that initializes the project with
// values from the specified key file and the cloud using the given project ID.
// Note, the entity identified by the key file passed into this function
// requires the resourcemanager.projects.get permission to be able to read the
// project name and project number.
func KeyFileWithProject(keyFile, projectID string) ProjectFunc {
	return func(ctx context.Context) (project, error) {
		creds, err := readCredentials(ctx, keyFile)
		if err != nil {
			return project{}, fmt.Errorf("failed to read credentials: %s", err)
		}

		return gcpProject(ctx, creds, projectID)
	}
}

// Project returns a ProjectFunc that initializes the project with the specified
// values. The project name is constructed from the project ID. The organization
// name is left blank.
func Project(projectID string, projectNumber int64) ProjectFunc {
	return func(ctx context.Context) (project, error) {
		name := cases.Title(language.Und).String(strings.ReplaceAll(projectID, "-", " "))
		return project{NativeID: projectID, number: projectNumber, name: name, orgName: ""}, nil
	}
}

// readCredentials reads the credentials from the specified key file.
// File related errors are not propagated since they might contain sensitive
// information, e.g. an accidentally malformed GCP service account key is passed
// to the Key function.
func readCredentials(ctx context.Context, keyFile string) (*google.Credentials, error) {
	// Expand the ~ token to the user's home directory. This should never fail
	// unless the shell environment is broken.
	if homeToken := fmt.Sprintf("~%c", filepath.Separator); strings.HasPrefix(keyFile, homeToken) {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home dir: %s", err)
		}
		keyFile = filepath.Join(home, strings.TrimPrefix(keyFile, homeToken))
	}

	// Stat the file to determine if it exists, stat doesn't require access
	// permissions on the file.
	if info, err := os.Stat(keyFile); err != nil || info.IsDir() {
		if err == nil {
			return nil, errors.New("key file is a directory")
		}
		return nil, errors.New("key file not found")
	}
	buf, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, errors.New("failed to read key file")
	}

	creds, err := google.CredentialsFromJSON(ctx, buf, scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain GCP credentials from key file: %s", err)
	}

	return creds, nil
}

// gcpProject returns a project initialized values from the credentials and the
// cloud.
// Note, the entity identified by the credentials passed into this function
// requires the resourcemanager.projects.get permission to be able to read the
// project name and project number.
func gcpProject(ctx context.Context, creds *google.Credentials, id string) (project, error) {
	if id == "" {
		return project{}, errors.New("project id cannot be empty")
	}

	client, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return project{}, fmt.Errorf("failed to create GCP Cloud Resource Manager client: %s", err)
	}

	// Lookup project.
	proj, err := client.Projects.Get(id).Do()
	if err != nil {
		return project{}, fmt.Errorf("failed to get GCP project: %s", err)
	}

	// Try to lookup parent organization's display name, ignoring errors.
	orgName := ""
	if proj.Parent != nil {
		if proj.Parent.Type == "organization" {
			if org, err := client.Organizations.Get("organizations/" + proj.Parent.Id).Do(); err == nil {
				if org.DisplayName != "" {
					orgName = org.DisplayName
				}
			}
		}
	}

	project := project{
		NativeID: id,
		name:     proj.Name,
		number:   proj.ProjectNumber,
		orgName:  orgName,
		creds:    creds,
	}

	return project, nil
}
