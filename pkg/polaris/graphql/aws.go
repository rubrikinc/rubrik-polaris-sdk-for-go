// Copyright 2021 Rubrik, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package graphql

// AwsCloudAccounts -
type AwsCloudAccount struct {
	AwsCloudAccount struct {
		CloudType           string `json:"cloudType"`
		ID                  string `json:"id"`
		NativeID            string `json:"nativeId"`
		AccountName         string `json:"accountName"`
		Message             string `json:"message"`
		SeamlessFlowEnabled bool   `json:"seamlessFlowEnabled"`
	} `json:"awsCloudAccount"`
	FeatureDetails []struct {
		AwsRegions []string `json:"awsRegions"`
		Feature    string   `json:"feature"`
		RoleArn    string   `json:"roleArn"`
		StackArn   string   `json:"stackArn"`
		Status     string   `json:"status"`
	} `json:"featureDetails"`
}

// AwsNativeAccount -
type AwsNativeAccount struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Regions       []string `json:"regions"`
	Status        string   `json:"status"`
	SLAAssignment string   `json:"slaAssignment"`

	ConfiguredSLADomain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"configuredSlaDomain"`

	EffectiveSLADomain struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"effectiveSlaDomain"`
}

// AwsNativeAccountConnection -
// type AwsNativeAccountConnection struct {
// 	Count int `json:"count"`
// 	Edges []struct {
// 		Node struct {
// 			ID            string   `json:"id"`
// 			Regions       []string `json:"regions"`
// 			Status        string   `json:"status"`
// 			Name          string   `json:"name"`
// 			SLAAssignment string   `json:"slaAssignment"`

// 			ConfiguredSLADomain struct {
// 				ID   string `json:"id"`
// 				Name string `json:"name"`
// 			} `json:"configuredSlaDomain"`

// 			EffectiveSLADomain struct {
// 				ID   string `json:"id"`
// 				Name string `json:"name"`
// 			} `json:"effectiveSlaDomain"`
// 		} `json:"node"`
// 	} `json:"edges"`
// 	PageInfo struct {
// 		EndCursor   string `json:"endCursor"`
// 		HasNextPage bool   `json:"hasNextPage"`
// 	} `json:"pageInfo"`
// }

// AwsNativeProtectionAccountAddResponse -
// type AwsNativeProtectionAccountAddResponse struct {
// 	CloudFormationName        string `json:"cloudFormationName"`
// 	CloudFormationTemplateURL string `json:"cloudFormationTemplateUrl"`
// 	CloudFormationURL         string `json:"cloudFormationUrl"`
// 	ErrorMessage              string `json:"errorMessage"`
// }
