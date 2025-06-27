package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/kr/pretty"
	"golang.org/x/sync/errgroup"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testsetup"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/aws"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/azure"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/exocompute"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/gcp"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/core"
	polaris_log "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

func main() {
	cleanup := flag.Bool("cleanup", false, "Perform cleanup tasks post CI")
	provider := flag.String("provider", "", "Use a specific cloud service provider: AWS, Azure or GCP")
	flag.Parse()

	// Load configuration and create a client
	polAccount, err := polaris.DefaultServiceAccount(true)
	if err != nil {
		log.Fatal(err)
	}
	logger := polaris_log.NewStandardLogger()
	logger.SetLogLevel(polaris_log.Info)
	if err := polaris.SetLogLevelFromEnv(logger); err != nil {
		log.Fatal(err)
	}
	client, err := polaris.NewClientWithLogger(polAccount, logger)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	err = check(ctx, client, strings.ToLower(*provider))
	switch {
	case err != nil && !*cleanup:
		log.Fatal(err)
	case err != nil:
		log.Print(err)
	}

	if *cleanup {
		if err := clean(ctx, client, strings.ToLower(*provider)); err != nil {
			log.Fatal(err)
		}
	}
}

func check(ctx context.Context, client *polaris.Client, provider string) error {
	var g errgroup.Group

	// AWS with profile
	if provider == "" || provider == "aws" {
		g.Go(func() error {
			testAcc, err := testsetup.AWSAccount()
			if err != nil {
				return err
			}
			awsAccount, err := aws.Wrap(client).Account(ctx, aws.AccountID(testAcc.AccountID), core.FeatureAll)
			switch {
			case err == nil:
				return fmt.Errorf("found pre-existing AWS account: %s\n%v", awsAccount.ID, pretty.Sprint(awsAccount))
			case !errors.Is(err, graphql.ErrNotFound):
				return fmt.Errorf("failed to check AWS account: %v", err)
			}
			return nil
		})

		// AWS with a cross-account role
		g.Go(func() error {
			testAcc, err := testsetup.AWSAccount()
			if err != nil {
				return err
			}
			awsAccount, err := aws.Wrap(client).Account(ctx, aws.AccountID(testAcc.CrossAccountID), core.FeatureAll)
			switch {
			case err == nil:
				return fmt.Errorf("found pre-existing AWS account: %s\n%v", awsAccount.ID, pretty.Sprint(awsAccount))
			case !errors.Is(err, graphql.ErrNotFound):
				return fmt.Errorf("failed to check AWS account: %v", err)
			}
			return nil
		})
	}

	// Azure
	if provider == "" || provider == "azure" {
		g.Go(func() error {
			testSub, err := testsetup.AzureSubscription()
			if err != nil {
				return err
			}
			azureAcc, err := azure.Wrap(client).Subscription(ctx, azure.SubscriptionID(testSub.SubscriptionID), core.FeatureAll)
			switch {
			case err == nil:
				return fmt.Errorf("found pre-existing Azure subscription: %s\n%v", azureAcc.ID, pretty.Sprint(azureAcc))
			case !errors.Is(err, graphql.ErrNotFound):
				return fmt.Errorf("failed to check Azure account: %v", err)
			}
			return nil
		})
	}

	// GCP
	if provider == "" || provider == "gcp" {
		g.Go(func() error {
			testProj, err := testsetup.GCPProject()
			if err != nil {
				return err
			}
			proj, err := gcp.Wrap(client).Project(ctx, gcp.ProjectID(testProj.ProjectID), core.FeatureAll)
			switch {
			case err == nil:
				return fmt.Errorf("found pre-existing GCP projects: %s\n%v", proj.ID, pretty.Sprint(proj))
			case !errors.Is(err, graphql.ErrNotFound):
				return fmt.Errorf("failed to check GCP project: %v", err)
			}
			return nil
		})
	}

	return g.Wait()
}

func clean(ctx context.Context, client *polaris.Client, provider string) error {
	var g errgroup.Group

	// AWS with profile
	if provider == "" || provider == "aws" {
		g.Go(func() error {
			testAcc, err := testsetup.AWSAccount()
			if err != nil {
				return err
			}

			awsClient := aws.Wrap(client)
			awsAccount, err := awsClient.Account(ctx, aws.AccountID(testAcc.AccountID), core.FeatureAll)
			switch {
			case errors.Is(err, graphql.ErrNotFound):
				return nil
			case err != nil:
				return fmt.Errorf("failed to check AWS account: %v", err)
			}
			if awsAccount.NativeID != testAcc.AccountID {
				return fmt.Errorf("existing AWS account %q isn't expected test account %q, won't remove",
					awsAccount.NativeID, testAcc.AccountID)
			}

			features := make([]core.Feature, 0, len(awsAccount.Features))
			for _, feature := range awsAccount.Features {
				features = append(features, feature.Feature)
			}
			return awsClient.RemoveAccount(ctx, aws.Profile(testAcc.Profile), features, false)
		})

		// AWS with a cross-account role
		g.Go(func() error {
			testAcc, err := testsetup.AWSAccount()
			if err != nil {
				return err
			}

			awsClient := aws.Wrap(client)
			awsAccount, err := awsClient.Account(ctx, aws.AccountID(testAcc.CrossAccountID), core.FeatureAll)
			switch {
			case errors.Is(err, graphql.ErrNotFound):
				return nil
			case err != nil:
				return fmt.Errorf("failed to check AWS account: %v", err)
			}
			if awsAccount.NativeID != testAcc.CrossAccountID {
				return fmt.Errorf("existing AWS account %q isn't expected test account %q, won't remove",
					awsAccount.NativeID, testAcc.CrossAccountID)
			}

			features := make([]core.Feature, 0, len(awsAccount.Features))
			for _, feature := range awsAccount.Features {
				features = append(features, feature.Feature)
			}
			return awsClient.RemoveAccount(ctx, aws.DefaultWithRole(testAcc.CrossAccountRole), features, false)
		})
	}

	// Azure
	if provider == "" || provider == "azure" {
		g.Go(func() error {
			testSub, err := testsetup.AzureSubscription()
			if err != nil {
				return err
			}

			azureClient := azure.Wrap(client)
			azureAcc, err := azureClient.Subscription(ctx, azure.SubscriptionID(testSub.SubscriptionID), core.FeatureAll)
			switch {
			case errors.Is(err, graphql.ErrNotFound):
				return nil
			case err != nil:
				return fmt.Errorf("failed to check Azure subscription: %v", err)
			}
			if azureAcc.NativeID != testSub.SubscriptionID {
				return fmt.Errorf("existing Azure subscription %q isn't the expected test subscription %q, won't remove",
					azureAcc.NativeID, testSub.SubscriptionID)
			}

			// Polaris doesn't automatically remove exocompute configs when removing
			// the subscription, so we need to do it manually here.
			exoClient := exocompute.Wrap(client)
			exoCfgs, err := exoClient.AzureConfigurationsByCloudAccountID(ctx, azureAcc.ID)
			if err != nil {
				return err
			}
			for i := range exoCfgs {
				if err := exoClient.RemoveAzureConfiguration(ctx, exoCfgs[i].ID); err != nil {
					return fmt.Errorf("failed to remove Azure ExocomputeConfig: %v", pretty.Sprint(exoCfgs[i]))
				}
			}

			// Remove all features for the subscription.
			for _, feature := range azureAcc.Features {
				if err := azureClient.RemoveSubscription(ctx, azure.CloudAccountID(azureAcc.ID), feature.Feature, false); err != nil {
					return fmt.Errorf("failed to remove Azure cloud account fetaure %v: %s", feature.Name, err)
				}
			}

			return nil
		})
	}

	// GCP
	if provider == "" || provider == "gcp" {
		g.Go(func() error {
			testProj, err := testsetup.GCPProject()
			if err != nil {
				return err
			}

			gcpClient := gcp.Wrap(client)
			proj, err := gcpClient.Project(ctx, gcp.ProjectID(testProj.ProjectID), core.FeatureAll)
			switch {
			case errors.Is(err, graphql.ErrNotFound):
				return nil
			case err != nil:
				return fmt.Errorf("failed to check GCP project: %v", err)
			}
			if pn := proj.ProjectNumber; pn != testProj.ProjectNumber {
				return fmt.Errorf("existing GCP project %q isn't expected test project %q, won't remove",
					pn, testProj.ProjectNumber)
			}

			return gcpClient.RemoveProject(ctx, gcp.ProjectNumber(testProj.ProjectNumber), core.FeatureCloudNativeProtection, false)
		})
	}

	return g.Wait()
}
