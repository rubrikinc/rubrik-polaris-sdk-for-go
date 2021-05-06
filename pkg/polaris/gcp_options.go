package polaris

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

// GcpConfigOption accepts GCP configuration options.
type GcpConfigOption interface {
	gcpConfig(ctx context.Context, opts *options) error
}

func gcpProject(ctx context.Context, creds *google.Credentials, projectID string, opts *options) error {
	if opts == nil {
		return errors.New("polaris: opts cannot be nil")
	}

	client, err := cloudresourcemanager.NewService(ctx, option.WithCredentials(creds))
	if err != nil {
		return err
	}

	opts.gcpCreds = creds

	// Lookup project.
	projReq := client.Projects.Get(projectID)
	projRes, err := projReq.Do()
	if err != nil {
		return err
	}

	opts.gcpName = projRes.Name
	opts.gcpID = projRes.ProjectId
	opts.gcpNumber = projRes.ProjectNumber

	// Lookup parent organization.
	orgName := projRes.Parent.Id
	if projRes.Parent.Type == "organization" {
		orgReq := client.Organizations.Get("organizations/" + projRes.Parent.Id)
		orgRes, err := orgReq.Do()
		if err != nil {
			return err
		}

		if orgRes.DisplayName != "" {
			orgName = orgRes.DisplayName
		}
	}

	opts.gcpOrgName = orgName
	return nil
}

func gcpCredentialsFromKeyFile(ctx context.Context, keyFile string) (*google.Credentials, error) {
	if strings.HasPrefix(keyFile, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		keyFile = filepath.Join(home, strings.TrimPrefix(keyFile, "~/"))
	}

	// Read GCP credentials. Note that a service account can be granted
	// access to projects other than the one it's created in.
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

type gcpConfigOption struct {
	parse func(context.Context, *options) error
}

func (o *gcpConfigOption) gcpConfig(ctx context.Context, opts *options) error {
	return o.parse(ctx, opts)
}

func (o *gcpConfigOption) query(ctx context.Context, opts *options) error {
	return o.parse(ctx, opts)
}

// FromGcpDefault passes the default GCP configuration as an option to a
// function accepting GcpConfigOption or QueryOption as argument. When given
// multiple times to a variadic function the last key file given will be used.
func FromGcpDefault() *gcpConfigOption {
	return &gcpConfigOption{func(ctx context.Context, opts *options) error {
		creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
		if err != nil {
			return err
		}

		return gcpProject(ctx, creds, creds.ProjectID, opts)
	}}
}

// FromGcpKeyFile passes the GCP configuration identified by the given key file
// as an option to a function accepting GcpConfigOption or QueryOption as
// argument. When given multiple times to a variadic function the last key file
// given will be used.
func FromGcpKeyFile(keyFile string) *gcpConfigOption {
	return &gcpConfigOption{func(ctx context.Context, opts *options) error {
		creds, err := gcpCredentialsFromKeyFile(ctx, keyFile)
		if err != nil {
			return err
		}

		return gcpProject(ctx, creds, creds.ProjectID, opts)
	}}
}

// FromGcpKeyFileWithProjectID passes the GCP configuration identified by the
// given key file and project id as an option to a function accepting
// GcpConfigOption or QueryOption as argument. When given multiple times to a
// variadic function the last key file given will be used.
func FromGcpKeyFileWithProjectID(keyFile, projectID string) *gcpConfigOption {
	return &gcpConfigOption{func(ctx context.Context, opts *options) error {
		creds, err := gcpCredentialsFromKeyFile(ctx, keyFile)
		if err != nil {
			return err
		}

		return gcpProject(ctx, creds, projectID, opts)
	}}
}

// FromGcpProject passes the GCP project details as an option to a function
// accepting GcpConfigOption or QueryOption as argument. When given multiple
// times to a variadic function the details given will be used.
func FromGcpProject(projectID, projectName string, projectNumber int64, orgName string) *gcpConfigOption {
	return &gcpConfigOption{func(ctx context.Context, opts *options) error {
		if opts == nil {
			return errors.New("polaris: opts cannot be nil")
		}

		opts.gcpCreds = nil
		opts.gcpID = projectID
		opts.gcpName = projectName
		opts.gcpNumber = projectNumber
		opts.gcpOrgName = orgName

		return nil
	}}
}

// WithGcpProjectID passes the specified GCP project id as an option to a
// function accepting QueryOption as argument. When given multiple times to a
// variadic function only the first project id will be used. Note that cloud
// service provider specific options that also specifies a project id, directly
// or indirectly, takes priority.
func WithGcpProjectID(projectID string) *queryOption {
	return &queryOption{func(ctx context.Context, opts *options) error {
		if opts == nil {
			return errors.New("polaris: opts cannot be nil")
		}

		if opts.gcpID == "" {
			opts.gcpID = projectID
		}

		return nil
	}}
}

// WithGcpProjectNumber passes the specified GCP project number as an option
// to a function accepting QueryOption as argument. When given multiple times
// to a variadic function only the first project number will be used. Note
// that cloud service provider specific options that also specifies a project
// number, directly or indirectly, takes priority.
func WithGcpProjectNumber(projectNumber int64) *queryOption {
	return &queryOption{func(ctx context.Context, opts *options) error {
		if opts == nil {
			return errors.New("polaris: opts cannot be nil")
		}

		if opts.gcpNumber == 0 {
			opts.gcpNumber = projectNumber
		}

		return nil
	}}
}
