package polaris

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
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

// AddOption accept options valid for an add operation.
type AddOption interface {
	add(ctx context.Context, opts *options) error
}

// IDOption accept options valid as id for an operation.
type IDOption interface {
	id(ctx context.Context, opts *options) error
}

// QueryOption accepts options valid for a query operation.
type QueryOption interface {
	query(ctx context.Context, opts *options) error
}

type addOption struct {
	parse func(ctx context.Context, opts *options) error
}

func (o *addOption) add(ctx context.Context, opts *options) error {
	return o.parse(ctx, opts)
}

// WithAddOption passes the specified region as an option to a function
// accepting AddOption as argument. When given multiple times to a variadic
// function all regions will be used.
func WithRegion(region string) *addOption {
	return &addOption{func(ctx context.Context, opts *options) error {
		for _, r := range opts.regions {
			if region == r {
				return nil
			}
		}

		opts.regions = append(opts.regions, region)
		return nil
	}}
}

// WithRegions passes the specified set of regions as an option to a function
// accepting AddOption as argument. When given multiple times to a variadic
// function all regions will be used.
func WithRegions(regions ...string) *addOption {
	return &addOption{func(ctx context.Context, opts *options) error {
		set := make(map[string]struct{}, len(regions)+len(opts.regions))
		for _, region := range opts.regions {
			set[region] = struct{}{}
		}

		// Add regions not already in the set.
		for _, region := range regions {
			if _, ok := set[region]; !ok {
				opts.regions = append(opts.regions, region)
				set[region] = struct{}{}
			}
		}

		return nil
	}}
}

type idOption struct {
	parse func(context.Context, *options) error
}

func (o *idOption) id(ctx context.Context, opts *options) error {
	return o.parse(ctx, opts)
}

// WithUUID passes the specified uuid as an option to a function accepting
// IDOption as an argument. When given multiple times to a variadic function
// the last uuid given will be used.
func WithUUID(id string) *idOption {
	return &idOption{func(ctx context.Context, opts *options) error {
		if _, err := uuid.Parse(id); err != nil {
			return err
		}

		opts.id = id
		return nil
	}}
}

type addAndQueryOption struct {
	parse func(ctx context.Context, opts *options) error
}

func (o *addAndQueryOption) add(ctx context.Context, opts *options) error {
	return o.parse(ctx, opts)
}

func (o *addAndQueryOption) query(ctx context.Context, opts *options) error {
	return o.parse(ctx, opts)
}

// WithName passes the specified name as an option to a function accepting
// AddOption or QueryOption as argument. When given multiple times to a
// variadic function the last name given will be used.
func WithName(name string) *addAndQueryOption {
	return &addAndQueryOption{func(ctx context.Context, opts *options) error {
		opts.name = name
		return nil
	}}
}
