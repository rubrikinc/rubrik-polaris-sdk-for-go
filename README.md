![Go version](https://img.shields.io/github/go-mod/go-version/rubrikinc/rubrik-polaris-sdk-for-go) ![License MIT](https://img.shields.io/github/license/rubrikinc/rubrik-polaris-sdk-for-go) ![Latest tag](https://img.shields.io/github/v/tag/rubrikinc/rubrik-polaris-sdk-for-go)

# Rubrik Polaris SDK for Go
Documentation for the SDK can be found [here](https://pkg.go.dev/github.com/rubrikinc/rubrik-polaris-sdk-for-go@main).
Please note that the repository has been repurposed so there exist older cached versions with a higher version number.

## Build the SDK
To build the SDK nothing except the standard Go build tools are required. To build all packages except tests run:
```
$ go build ./...
```

To build one of the examples run:
```
$ go build ./examples/<example-to-build>
```

To transform new or updated GraphQL queries into Go code run:
```
$ go generate ./...
```

## Run Code Using the SDK

### Environment Variables
The following environmental variables can be used to override the default behaviour of the SDK:
* *RUBRIK_POLARIS_LOGLEVEL* — Overrides the log level of the SDK. Valid log levels are: *FATAL*, *ERROR*, *WARN*,
  *INFO*, *DEBUG*, *TRACE* and *OFF*. The default log level is *WARN*.
* *RUBRIK_POLARIS_TOKEN_CACHE* — Overrides whether the token cache should be used or not.
* *RUBRIK_POLARIS_TOKEN_CACHE_DIR* — Overrides the directory where cached authentication tokens are stored.
* *RUBRIK_POLARIS_TOKEN_CACHE_SECRET* — Overrides the secret used as input when generating an encryption key for the
  authentication token.

Note that it's possible to prevent the above environment variables, except for *RUBRIK_POLARIS_LOGLEVEL*, from
overriding the default behavior by setting `allowEnvOverride` to `false` for the account passed in when creating the
client.

### Polaris Credentials
The SDK supports both local user accounts and service accounts. For documentation on how to create either using Polaris
see the [Rubrik Support Portal](http://support.rubrik.com/). The recommendation is to always use a service account with
the SDK. 

#### Local User Account
To use a local user account with the SDK first create a directory called `.rubrik` in your home directory. In that
directory create a file called `polaris-accounts.json`. This JSON file can hold one or more local user accounts as per
this pattern:
```
{
    "<my-account>": {
        "username": "<my-username>",
        "password": "<my-password>",
        "url": "<my-polaris-url>"
    }
}
```
Where `my-account` is an arbitrary name used to refer to the account when initializing the SDK. `my-username` and
`my-password` are the username and password of the local user account. `my-polaris-url` is the URL of the Polaris API.
The API URL normally follows the pattern `https://{polaris-domain}.my.rubrik.com/api`. Which is the same URL as for
accessing the Polaris UI but with `/api` added to the end.

As an example, assume our Polaris domain is `my-polaris-domain` and that the username and password of our local user
account is `john.doe@example.org` and `password123` the content of the `polaris-accounts.json` file then should be:
```
{
    "johndoe": {
        "username": "john.doe@example.org",
        "password": "password123",
        "url": "https://my-polaris-domain.my.rubrik.com/api"
    }
}
```
Where `johndoe` will be used to refer to this account when initializing the SDK:
```
account, err := polaris.DefaultUserAccount("johndoe", true)
if err != nil  {
    log.Fatal(err)
}

client, err := polaris.NewClient(context.Background(), account, polaris_log.NewStandardLogger())
if err != nil  {
    log.Fatal(err)
}
```
Two additional functions exist to use a local user account with Polaris: `polaris.UserAccountFromFile` and 
`polaris.UserAccountFromEnv`. Please see the
[SDK documentation](https://pkg.go.dev/github.com/rubrikinc/rubrik-polaris-sdk-for-go@main/pkg/polaris#UserAccount) for
details on how to use them.

#### Local User Account Environment Variables
When using a local user account the following environmental variables can be used to override the default local user
account behaviour:
* *RUBRIK_POLARIS_ACCOUNT_FILE* — Overrides the name and path of the file to read local user accounts from.
* *RUBRIK_POLARIS_ACCOUNT_NAME* — Overrides the name of the local user account given to the SDK during initialization.
* *RUBRIK_POLARIS_ACCOUNT_USERNAME* — Overrides the username of the local user account.
* *RUBRIK_POLARIS_ACCOUNT_PASSWORD* — Overrides the password of the local user account.
* *RUBRIK_POLARIS_ACCOUNT_URL* — Overrides the Polaris API URL.

Note that it's possible to prevent the above environment variables from overriding the default behavior by setting
`allowEnvOverride` to `false`.

#### Service Account
To use a service account with the SDK first create a directory called `.rubrik` in your home directory. Next, download
the service account credentials from the Polaris user management page to a file in that directory named 
`polaris-service-account.json`. The `polaris-service-account.json` file contains everything needed to connect to
Polaris from the SDK:
```
account, err := polaris.DefaultServiceAccount(true)
if err != nil  {
    log.Fatal(err)
}

client, err := polaris.NewClient(context.Background(), account, polaris_log.NewStandardLogger())
if err != nil  {
    log.Fatal(err)
}
```
Two additional functions exist to use a service account with Polaris: `polaris.ServiceAccountFromFile` and
`polaris.ServiceAccountFromEnv`. Please see the
[SDK documentation](https://pkg.go.dev/github.com/rubrikinc/rubrik-polaris-sdk-for-go@main/pkg/polaris#ServiceAccount)
for details on how to use them.

#### Service Account Environment Variables
When using a service account the following environmental variables can be used to override the default service account
behavior:
* *RUBRIK_POLARIS_SERVICEACCOUNT_FILE* — Overrides the name and path of the service account credentials file.
* *RUBRIK_POLARIS_SERVICEACCOUNT_NAME* — Overrides the name of the service account.
* *RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTID* — Overrides the client id of the service account.
* *RUBRIK_POLARIS_SERVICEACCOUNT_CLIENTSECRET* — Overrides the client secret of the service account.
* *RUBRIK_POLARIS_SERVICEACCOUNT_ACCESSTOKENURI* — Overrides the service account access token URI. When using a service 
account the Polaris API URL is derived from this URI.

Note that it's possible to prevent the above environment variables from overriding the default behavior by setting
`allowEnvOverride` to `false`.

### AWS Credentials
To perform AWS operations with the SDK an AWS profile and region is required. The SDK will look for those in the
default `~/.aws/credentials` and `~/.aws/config` files. The profile and region used by the SDK depends on the account
function used, please see the
[SDK documentation](https://pkg.go.dev/github.com/rubrikinc/rubrik-polaris-sdk-for-go@main/pkg/polaris/aws#AccountFunc)
for more information.

### Azure Credentials
To perform Azure operations with the SDK an Azure service principal is required. Service principals are referred to as
app registrations in the Azure portal. The easiest way to create a service principal is to use the Azure CLI tool:
```
$ az ad sp create-for-rbac --sdk-auth=true -n "<my-app-name>"
```
This will create an Azure service principal and output a JSON snippet that can be used to authenticate as the service
principal to Azure. Note that you might need to specify `--role` and `--scopes` to give your service principal the
required Azure permissions. The SDK will look for the JSON snippet in the file pointed to by the `AZURE_AUTH_LOCATION`
environment variable if the service principal function `azure.Default` is used. The file can also be pointed out
directly using the service principal function `azure.SDKAuthFile`. The service principal can also be given as parameters
using the `azure.ServicePrincipal` function.

If the service principal is created using the Azure portal there will be no JSON snippet, instead the detailed app
registration information can be copied from the portal to a JSON file having the following structure:
```
{
    "appId": "<app-id>",
    "appName": "<app-name>",
    "appSecret": "<app-secret>",
    "tenantId": "<tenant-id>"
}
```
The SDK can be pointed to this file by using the `KeyFile` service principal function. Please see the
[SDK documentation](https://pkg.go.dev/github.com/rubrikinc/rubrik-polaris-sdk-for-go@main/pkg/polaris/azure#ServicePrincipalFunc)
for more information.

### GCP Credentials
To perform GCP operations with the SDK a GCP service account is required. The SDK will look for the service account in
the file pointed to by the `GOOGLE_APPLICATION_CREDENTIALS` environment variable if the project function `gcp.Default`
is used. The file can also be pointed out directly using the project functions `gcp.KeyFile` and
`gcp.KeyFileAndProject`. Please see the
[SDK documentation](https://pkg.go.dev/github.com/rubrikinc/rubrik-polaris-sdk-for-go@main/pkg/polaris/gcp#ProjectFunc)
for more information.

## Run the SDK Test Suite

### Unit Tests
To execute the unit test suite run:
```
$ go test ./...
```

### Integration Tests
Note that the integration tests requires an RSC instance and, depending on which tests are run, an appliance
connected to the Polaris instance, an AWS account, an Azure subscription and a GCP project. See below for additional
requirements for each cloud service provider.

To execute the integration test suite run:
```
$ TEST_INTEGRATION=1 go test -timeout=60m ./...
```

#### Access
To run the access integration tests, an RSC test user must be created. It also requires that the environment variable
`TEST_RSCCONFIG_FILE` points to a JSON file containing information used to assert that users and access operation were
performed correctly:
```json
{
    "existingUserEmail": "<existing-rsc-test-user-email-address>",
    "newUserEmail": "<non-existing-rsc-test-user-email-address>"
}
```

#### Appliance
To run the appliance token exchange integration test, an appliance/cluster must already be registered to the Polaris
instance and the environment variable `TEST_APPLIANCE_ID` must be set to the id (UUID) of the registered cluster.

#### AWS
Requires a default AWS profile along with a default region for the profile. It also requires that the environment
variable `TEST_AWSACCOUNT_FILE` points to a JSON file containing information used to assert that the account was added
correctly to Polaris:
```json
{
    "profile": "<aws-profile-name>",
    "accountId": "<aws-account-id>",
    "accountName": "<aws-account-name>",
    "crossAccountId": "<aws-cross-account-id>",
    "crossAccountName": "<aws-cross-account-name>",
    "crossAccountRole": "<aws-cross-account-role>",
    "exocompute": {
        "vpcId": "<aws-vpc-id>",
        "subnets": [{
            "id": "<aws-subnet-id>",
            "availabilityZone": "<aws-availability-zone>"
        }, {
            "id": "<aws-subnet-id>",
            "availabilityZone": "<aws-availability-zone>"
        }, {
            "id": "<aws-subnet-id>",
            "availabilityZone": "<aws-availability-zone>"
        }]
    }
}
```
Note that the exocompute part is only needed when running the AWS Exocompute integration test.

#### Azure
To run the Azure integration tests an Azure service principal is required. It also requires that the environment
variable `TEST_AZURESUBSCRIPTION_FILE` points to a JSON file containing information used to assert that the account was
added correctly to Polaris:
```json
{
    "subscriptionId": "<azure-subscription-id>",
    "subscriptionName": "<azure-subscription-name>",
    "tenantId": "<azure-tenant-id>",
    "tenantDomain": "<azure-tenant-domain>",
    "principalId": "<azure-principal-id>",
    "principalName": "<azure-principal-name>",
    "principalSecret": "<azure-principal-secret>",
    "exocompute": {
        "subnetId": "<azure-subnet-id>"
    }
}
```
Note that the exocompute part is only needed when running the Azure Exocompute integration test.

#### GCP
To run the GCP integration tests a GCP service account is required. It also requires that the environment
variable `TEST_GCPPROJECT_FILE` points to a JSON file containing information used to assert that the account was added
correctly to Polaris:
```json
{
    "projectId": "<gcp-project-id>",
    "projectName": "<gcp-project-name>",
    "projectNumber": <gcp-project-number>,
    "organizationName": "<gcp-organization-name>"
}
```

## Request Appliance Token for Rubrik SDK for Go
To access Appliance REST APIs using Polaris service accounts, following are the prerequisites:
- Appliance/Cluster must be registered with the Polaris Instance.
- Appliance ID. Can be found under *Polaris Instance* -> *Clusters* > *\[Select a Cluster\]* -> *Cluster details* ->
*Id*.
- Appliance fully qualified domain name.

The SDK can be used to retrieve a token in exchange for the Polaris service account credentials and can be used
to access the appliance/cluster REST APIs. This token is only valid for a day. It is recommended to create a helper
method that auto refreshes token every 24 hours for long-running applications.
```
func ApplianceTokenExample(serviceAccountPath string, applianceID uuid.UUID, logger log.Logger) (string, error) {
    serviceAccount, err := polaris.ServiceAccountFromFile(serviceAccountPath, true)
    if err != nil {
        return "", err
    }
    
    token, err := appliance.TokenFromServiceAccount(serviceAccount, applianceID, logger)
    if err != nil {
        return "", err
    }

    return token, nil
}
```

The retrieved token can be used with the [rubrik-sdk-for-go](https://github.com/rubrikinc/rubrik-sdk-for-go) to set up a
CDM client and call CDM APIs.
```
func CallApplianceAPIExample(applianceFQDN, applianceToken string) {
    rubrik, err := rubrikcdm.ConnectAPIToken(applianceFQDN, applianceToken)
    if err != nil {
        log.Fatal(err)
    }
    
    // GET the Rubrik cluster Version
    clusterSummary, err := rubrik.Get("v1", "/cluster/me")
    if err != nil {
        log.Fatal(err)
    }
}
```

![48332236-55506f00-e610-11e8-9a60-594de963a1ee](https://user-images.githubusercontent.com/2046831/119498600-1580ab80-bd66-11eb-87ca-c08df4eae15a.png)
