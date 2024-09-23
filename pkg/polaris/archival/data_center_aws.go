package archival

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/archival"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// AWSCloudAccountByID returns the AWS data center cloud account with the
// specified ID. If no cloud account with the specified ID is found,
// graphql.ErrNotFound is returned.
func (a API) AWSCloudAccountByID(ctx context.Context, cloudAccountID uuid.UUID) (archival.AWSCloudAccount, error) {
	a.log.Print(log.Trace)

	cloudAccounts, err := a.AWSCloudAccounts(ctx, "")
	if err != nil {
		return archival.AWSCloudAccount{}, err
	}

	for _, account := range cloudAccounts {
		if account.ID == cloudAccountID {
			return account, nil
		}
	}

	return archival.AWSCloudAccount{}, fmt.Errorf("aws data center cloud account %q %w", cloudAccountID, graphql.ErrNotFound)
}

// AWSCloudAccountByName returns the AWS data center cloud account with the
// specified name. If no cloud account with the specified name is found,
// graphql.ErrNotFound is returned.
func (a API) AWSCloudAccountByName(ctx context.Context, name string) (archival.AWSCloudAccount, error) {
	a.log.Print(log.Trace)

	cloudAccounts, err := a.AWSCloudAccounts(ctx, name)
	if err != nil {
		return archival.AWSCloudAccount{}, err
	}

	name = strings.ToLower(name)
	for _, account := range cloudAccounts {
		if strings.ToLower(account.Name) == name {
			return account, nil
		}
	}

	return archival.AWSCloudAccount{}, fmt.Errorf("aws data center cloud account %q %w", name, graphql.ErrNotFound)
}

// AWSCloudAccounts returns all AWS data center cloud accounts matching the
// specified name filter. The name filter can be used to search for prefixes
// of a name. If the name filter is empty, it will match all names.
func (a API) AWSCloudAccounts(ctx context.Context, nameFilter string) ([]archival.AWSCloudAccount, error) {
	a.log.Print(log.Trace)

	var filters []archival.ListAccountFilter
	if nameFilter != "" {
		filters = append(filters, archival.ListAccountFilter{
			Field: "NAME",
			Text:  nameFilter,
		})
	}

	cloudAccounts, err := archival.ListAccounts[archival.AWSCloudAccount](ctx, a.client, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get aws data center cloud accounts: %s", err)
	}

	return cloudAccounts, nil
}

// CreateAWSCloudAccount creates a new AWS data center cloud account.
func (a API) CreateAWSCloudAccount(ctx context.Context, createParams archival.CreateAWSCloudAccountParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	accountID, err := archival.CreateCloudAccount[archival.CreateAWSCloudAccountResult](ctx, a.client, createParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create aws data center cloud account: %s", err)
	}

	return accountID, nil
}

// UpdateAWSCloudAccount updates the AWS data center cloud account with the
// specified cloud account ID.
func (a API) UpdateAWSCloudAccount(ctx context.Context, cloudAccountID uuid.UUID, updateParams archival.UpdateAWSCloudAccountParams) error {
	a.log.Print(log.Trace)

	if err := archival.UpdateCloudAccount[archival.UpdateAWSCloudAccountResult](ctx, a.client, cloudAccountID, updateParams); err != nil {
		return fmt.Errorf("failed to update aws data center cloud account: %s", err)
	}

	return nil
}

// DeleteAWSCloudAccount deletes the AWS data center cloud account with the
// specified cloud account ID.
func (a API) DeleteAWSCloudAccount(ctx context.Context, accountID uuid.UUID) error {
	a.log.Print(log.Trace)

	if err := archival.DeleteCloudAccount[archival.DeleteAWSCloudAccountResult](ctx, a.client, accountID); err != nil {
		return fmt.Errorf("failed to delete aws data center cloud account: %s", err)
	}

	return nil
}

// AWSTargetByID returns the AWS target with the specified target ID.
// If no target with the specified ID is found, graphql.ErrNotFound is returned.
func (a API) AWSTargetByID(ctx context.Context, targetID uuid.UUID) (archival.AWSTarget, error) {
	a.log.Print(log.Trace)

	filter := []archival.ListTargetFilter{{
		Field: "LOCATION_ID",
		Text:  targetID.String(),
	}}
	targets, err := archival.ListTargets[archival.AWSTarget](ctx, a.client, filter)
	if err != nil {
		return archival.AWSTarget{}, fmt.Errorf("failed to get targets: %s", err)
	}

	for _, target := range targets {
		if target.ID == targetID {
			return target, nil
		}
	}

	return archival.AWSTarget{}, fmt.Errorf("target for %q %w", targetID, graphql.ErrNotFound)
}

// AWSTargetByName returns the AWS target with the specified name.
// If no target with the specified name is found, graphql.ErrNotFound is
// returned.
func (a API) AWSTargetByName(ctx context.Context, name string) (archival.AWSTarget, error) {
	a.log.Print(log.Trace)

	targets, err := a.AWSTargets(ctx, name)
	if err != nil {
		return archival.AWSTarget{}, err
	}

	name = strings.ToLower(name)
	for _, target := range targets {
		if target.Name == name {
			return target, nil
		}
	}

	return archival.AWSTarget{}, fmt.Errorf("target for %q %w", name, graphql.ErrNotFound)
}

// AWSTargets returns all AWS targets matching the specified name filter.
// The name filter can be used to search for prefixes of a name. If the name
// filter is empty, is will match all names.
func (a API) AWSTargets(ctx context.Context, nameFilter string) ([]archival.AWSTarget, error) {
	a.log.Print(log.Trace)

	var filter []archival.ListTargetFilter
	if nameFilter != "" {
		filter = append(filter, archival.ListTargetFilter{
			Field: "NAME",
			Text:  nameFilter,
		})
	}
	targets, err := archival.ListTargets[archival.AWSTarget](ctx, a.client, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get targets: %s", err)
	}

	return targets, nil
}

// CreateAWSTarget creates a new AWS target.
func (a API) CreateAWSTarget(ctx context.Context, createParams archival.CreateAWSTargetParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	targetID, err := archival.CreateTarget[archival.CreateAWSTargetResult](ctx, a.client, createParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create AWS target: %s", err)
	}

	return targetID, nil
}

// UpdateAWSTarget updates the AWS target with the specified target ID.
func (a API) UpdateAWSTarget(ctx context.Context, targetID uuid.UUID, updateParams archival.UpdateAWSTargetParams) error {
	a.log.Print(log.Trace)

	if err := archival.UpdateTarget[archival.UpdateAWSTargetResult](ctx, a.client, targetID, updateParams); err != nil {
		return fmt.Errorf("failed update AWS target: %s", err)
	}

	return nil
}
