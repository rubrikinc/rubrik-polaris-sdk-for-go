package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polarislog "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func main() {
	ctx := context.Background()

	// Load configuration and create client.
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClientWithLogger(polAccount, polarislog.NewStandardLogger())
	if err != nil {
		log.Fatal(err)
	}

	awsClient := aws.Wrap(client.GQL)

	trustPolicies, err := awsClient.TrustPolicy(ctx, aws.CloudStandard, []core.Feature{core.FeatureCloudNativeProtection}, []aws.TrustPolicyAccount{{
		ID:         "311033699123",
		ExternalID: "",
	}})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Trust Policies: %+v\n", trustPolicies)
}
