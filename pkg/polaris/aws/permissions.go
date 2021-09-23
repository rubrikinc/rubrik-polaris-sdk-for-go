package aws

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// UpdatePermissions updates the permissions of the CloudFormation stack in
// AWS.
func (a API) UpdatePermissions(ctx context.Context, account AccountFunc, features []core.Feature) error {
	a.gql.Log().Print(log.Trace, "polaris/aws.UpdatePermissions")

	if account == nil {
		return errors.New("polaris: account is not allowed to be nil")
	}
	config, err := account(ctx)
	if err != nil {
		return err
	}

	akkount, err := a.Account(ctx, AccountID(config.id), core.FeatureAll)
	if err != nil {
		return err
	}

	cfmURL, tmplURL, err := aws.Wrap(a.gql).PrepareFeatureUpdateForAwsCloudAccount(ctx, akkount.ID, features)
	if err != nil {
		return err
	}

	// Extract stack id/name from returned CloudFormationURL.
	i := strings.LastIndex(cfmURL, "#/stack/update") + 1
	if i == 0 {
		return errors.New("polaris: CloudFormation url does not contain #/stack/detail")
	}

	u, err := url.Parse(cfmURL[i:])
	if err != nil {
		return err
	}
	stackID := u.Query().Get("stackId")

	err = awsUpdateStack(ctx, a.gql.Log(), config.config, stackID, tmplURL)
	if err != nil {
		return err
	}

	return nil
}
