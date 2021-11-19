package aws

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// UpdatePermissions updates the permissions of the CloudFormation stack in
// AWS.
func (a API) UpdatePermissions(ctx context.Context, account AccountFunc, features []core.Feature) error {
	a.gql.Log().Print(log.Trace)

	if account == nil {
		return errors.New("account is not allowed to be nil")
	}
	config, err := account(ctx)
	if err != nil {
		return fmt.Errorf("failed to lookup account: %v", err)
	}

	akkount, err := a.Account(ctx, AccountID(config.id), core.FeatureAll)
	if err != nil {
		return fmt.Errorf("failed to get account: %v", err)
	}

	cfmURL, tmplURL, err := aws.Wrap(a.gql).PrepareFeatureUpdateForAwsCloudAccount(ctx, akkount.ID, features)
	if err != nil {
		return fmt.Errorf("failed to update account: %v", err)
	}

	// Extract stack id/name from returned CloudFormationURL.
	i := strings.LastIndex(cfmURL, "#/stack/update") + 1
	if i == 0 {
		return errors.New("CloudFormation url does not contain #/stack/detail")
	}

	u, err := url.Parse(cfmURL[i:])
	if err != nil {
		return fmt.Errorf("failed to parse CloudFormation url: %v", err)
	}
	stackID := u.Query().Get("stackId")

	err = awsUpdateStack(ctx, a.gql.Log(), config.config, stackID, tmplURL)
	if err != nil {
		return fmt.Errorf("failed to update CloudFormation stack: %v", err)
	}

	return nil
}
