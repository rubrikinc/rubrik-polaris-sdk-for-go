package polaris

import (
	"context"
	"errors"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/uuid"
)

// options hold the combined result of all options passed as arguments to a
// function.
type options struct {
	id      string
	name    string
	regions []string

	// AWS specific options.
	awsID      string
	awsProfile string
	awsConfig  *aws.Config
}

// FromAwsOption accepts an AWS specific option as an argument.
type FromAwsOption struct {
	opt func(context.Context, *options) error
}

func (o *FromAwsOption) parse(ctx context.Context, opts *options) error {
	return o.opt(ctx, opts)
}

// WithOption accepts a generic option as an argument.
type WithOption struct {
	opt func(context.Context, *options) error
}

func (o *WithOption) parse(ctx context.Context, opts *options) error {
	return o.opt(ctx, opts)
}

// FromAwsOrWithOption accepts both an AWS specific option and a generic option
// as an argument.
type FromAwsOrWithOption interface {
	parse(context.Context, *options) error
}

// awsAccount returns the AWS account id and name. Note that if the AWS user
// does not have permissions for Organizations the account name will be empty.
func awsAccount(ctx context.Context, config aws.Config) (string, string, error) {
	stsClient := sts.NewFromConfig(config)
	id, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", "", err
	}

	// Organizations calls might fail due to missing permissions.
	orgClient := organizations.NewFromConfig(config)
	info, err := orgClient.DescribeAccount(ctx, &organizations.DescribeAccountInput{AccountId: id.Account})
	if err != nil {
		return *id.Account, "", nil
	}

	return *id.Account, *info.Account.Name, nil
}

// FromAwsConfig passes the specified AWS configuration as an option to a
// function accepting an FromAwsOption as an argument. When given multiple
// times to a variadic function the last name given will be used.
func FromAwsConfig(config aws.Config) *FromAwsOption {
	return &FromAwsOption{func(ctx context.Context, opts *options) error {
		id, name, err := awsAccount(ctx, config)
		if err != nil {
			return err
		}

		opts.awsID = id
		if name != "" {
			opts.name = name
		}

		opts.awsConfig = &config
		return nil
	}}
}

// FromAwsProfile passes the AWS configuration identified by the given
// profile as an option to a function accepting FromAwsOption as an argument.
// When given multiple times to a variadic function the last name given will
// be used.
func FromAwsProfile(profile string) *FromAwsOption {
	return &FromAwsOption{func(ctx context.Context, opts *options) error {
		config, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
		if err != nil {
			return err
		}

		id, name, err := awsAccount(ctx, config)
		if err != nil {
			return err
		}

		opts.awsID = id
		if name != "" {
			opts.name = name
		}

		opts.awsConfig = &config
		opts.awsProfile = profile
		return nil
	}}
}

// WithAwsID passes the specified id as an option to a function accepting
// WithOption as an argument. When given multiple times to a variadic function
// only the first id will be used. Note that cloud service provider specific
// options that also specifies an id, directly or indirectly, takes priority.
func WithAwsID(id string) *WithOption {
	return &WithOption{func(ctx context.Context, opts *options) error {
		if len(id) != 12 {
			return errors.New("polaris: invalid length for aws account id")
		}
		if opts.awsID == "" {
			opts.awsID = id
		}
		return nil
	}}
}

// WithUUID passes the specified uuid as an option to a function accepting
// WithOption as an argument. When given multiple times to a variadic function
// the last uuid given will be used.
func WithUUID(id string) *WithOption {
	return &WithOption{func(ctx context.Context, opts *options) error {
		if _, err := uuid.Parse(id); err != nil {
			return err
		}

		opts.id = id
		return nil
	}}
}

// WithName passes the specified name as an option to a function accepting
// WithOption as argument. When given multiple times to a variadic function
// the last name given will be used.
func WithName(name string) *WithOption {
	return &WithOption{func(ctx context.Context, opts *options) error {
		opts.name = name
		return nil
	}}
}

// WithRegions passes the specified set of regions as an option to a function
// accepting WithOption as argument. When given multiple times to a variadic
// function the last set of regions will be used.
func WithRegions(regions ...string) *WithOption {
	return &WithOption{func(ctx context.Context, opts *options) error {
		set := make(map[string]struct{}, len(regions))
		for _, region := range regions {
			set[region] = struct{}{}
		}

		opts.regions = make([]string, 0, len(set))
		for region := range set {
			opts.regions = append(opts.regions, region)
		}

		sort.Strings(opts.regions)
		return nil
	}}
}
