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

// AzureCloudAccountByID returns the Azure data center cloud account with the
// specified ID. If no cloud account with the specified ID is found,
// graphql.ErrNotFound is returned.
func (a API) AzureCloudAccountByID(ctx context.Context, cloudAccountID uuid.UUID) (archival.AzureCloudAccount, error) {
	a.log.Print(log.Trace)

	cloudAccounts, err := a.AzureCloudAccounts(ctx, "")
	if err != nil {
		return archival.AzureCloudAccount{}, err
	}

	for _, account := range cloudAccounts {
		if account.ID == cloudAccountID {
			return account, nil
		}
	}

	return archival.AzureCloudAccount{}, fmt.Errorf("azure data center cloud account %q %w", cloudAccountID, graphql.ErrNotFound)
}

// AzureCloudAccountByName returns the Azure data center cloud account with
// the specified name. If no cloud account with the specified name is found,
// graphql.ErrNotFound is returned.
func (a API) AzureCloudAccountByName(ctx context.Context, name string) (archival.AzureCloudAccount, error) {
	a.log.Print(log.Trace)

	cloudAccounts, err := a.AzureCloudAccounts(ctx, name)
	if err != nil {
		return archival.AzureCloudAccount{}, err
	}

	name = strings.ToLower(name)
	for _, account := range cloudAccounts {
		if strings.ToLower(account.Name) == name {
			return account, nil
		}
	}

	return archival.AzureCloudAccount{}, fmt.Errorf("azure data center cloud account %q %w", name, graphql.ErrNotFound)
}

// AzureCloudAccounts returns all Azure data center cloud accounts matching the
// specified name filter. The name filter can be used to search for prefixes
// of a name. If the name filter is empty, it will match all names.
func (a API) AzureCloudAccounts(ctx context.Context, nameFilter string) ([]archival.AzureCloudAccount, error) {
	a.log.Print(log.Trace)

	var filters []archival.ListAccountFilter
	if nameFilter != "" {
		filters = append(filters, archival.ListAccountFilter{
			Field: "NAME",
			Text:  nameFilter,
		})
	}

	cloudAccounts, err := archival.ListAccounts[archival.AzureCloudAccount](ctx, a.client, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get azure data center cloud accounts: %s", err)
	}

	return cloudAccounts, nil
}

// CreateAzureCloudAccount creates a new Azure data center cloud account.
func (a API) CreateAzureCloudAccount(ctx context.Context, createParams archival.CreateAzureCloudAccountParams) (uuid.UUID, error) {
	a.log.Print(log.Trace)

	accountID, err := archival.CreateCloudAccount[archival.CreateAzureCloudAccountResult](ctx, a.client, createParams)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create azure data center cloud account: %s", err)
	}

	return accountID, nil
}

// UpdateAzureCloudAccount updates the Azure data center cloud account with the
// specified cloud account ID.
func (a API) UpdateAzureCloudAccount(ctx context.Context, cloudAccountID uuid.UUID, updateParams archival.UpdateAzureCloudAccountParams) error {
	a.log.Print(log.Trace)

	if err := archival.UpdateCloudAccount[archival.UpdateAzureCloudAccountResult](ctx, a.client, cloudAccountID, updateParams); err != nil {
		return fmt.Errorf("failed to update azure data center cloud account: %s", err)
	}

	return nil
}

// DeleteAzureCloudAccount deletes the Azure data center cloud account with the
// specified cloud account ID.
func (a API) DeleteAzureCloudAccount(ctx context.Context, accountID uuid.UUID) error {
	a.log.Print(log.Trace)

	if err := archival.DeleteCloudAccount[archival.DeleteAzureCloudAccountResult](ctx, a.client, accountID); err != nil {
		return fmt.Errorf("failed to delete azure data center cloud account: %s", err)
	}

	return nil
}
