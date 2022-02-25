// Code generated by queries_gen.go DO NOT EDIT

// MIT License
//
// Copyright (c) 2021 Rubrik
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package k8s

// allSnapshotPvcs GraphQL query
var allSnapshotPvcsQuery = `query SdkGolangAllSnapshotPvcs(
    $snapshotId: String!,
    $snappableId: UUID!,
) {
    allSnapshotPvcs(snapshotId: $snapshotId, snappableId: $snappableId) {
        id
        name
        capacity
        accessMode
        storageClass
        volume
        labels
        phase
    }
}`

// getActivitySeries GraphQL query
var getActivitySeriesQuery = `query SdkGolangGetActivitySeries(
    $activitySeriesId: UUID!,
    $clusterUuid: UUID,
    $after: String,
) {
    activitySeries(activitySeriesId: $activitySeriesId, clusterUuid: $clusterUuid) {
        activityConnection(after: $after) {
            nodes {
                activityInfo
                message
                status
                time
                severity
            }
            pageInfo {
                endCursor
                hasNextPage
            }
            count
        }
    }
}`

// getActivitySeriesConnection GraphQL query
var getActivitySeriesConnectionQuery = `query SdkGolangGetActivitySeriesConnection(
    $after: String,
    $filters: ActivitySeriesFilterInput,
) {
    activitySeriesConnection(
        after: $after,
        filters: $filters,
    ) {
        edges {
            node {
                id
                lastActivityType
                lastActivityStatus
                severity
                objectId
                objectName
                objectType
                activitySeriesId
                progress
                activityConnection {
                    nodes {
                        message
                    }
                }
            }
        }
        pageInfo {
            endCursor,
            hasNextPage,
        },
        count,
    }
}`

// getNamespaces GraphQL query
var getNamespacesQuery = `query SdkGolangGetNamespaces(
    $after: String,
    $filter: [Filter!],
    $k8sClusterId: UUID,
) {
    k8sNamespaces(
        after: $after,
        filter: $filter,
        k8sClusterId: $k8sClusterId
    ) {
        edges {
            node {
                id,
                k8sClusterID,
                namespaceName,
                isRelic,
                configuredSlaDomain{
                    id,
                    name,
                    version,
                },
                effectiveSlaDomain{
                    id,
                    name,
                    version,
                },
            }
        },
        pageInfo {
            endCursor,
            hasNextPage,
        },
        count
    }
}`

// getTaskchainInfo GraphQL query
var getTaskchainInfoQuery = `query SdkGolangGetTaskchainInfo(
    $taskchainId: String!,
    $jobType: String!,
) {
    getTaskchainInfo(
        taskchainId: $taskchainId,
        jobType: $jobType,
    ) {
        taskchainId,
        state,
        startTime,
        endTime,
        progress,
        jobId,
        jobType,
        error,
        account,
    }
}`

// k8sNamespace GraphQL query
var k8sNamespaceQuery = `query SdkGolangK8sNamespace(
    $after: String
    $filter: PolarisSnapshotFilterInput,
    $fid: UUID!,
) {
    k8sNamespace(fid: $fid) {
        snapshotConnection(after: $after, filter: $filter) {
            pageInfo {
                endCursor
                hasNextPage
            }
            nodes {
                id
                date
                isOnDemandSnapshot
                expirationDate
                isCorrupted
                isDeletedFromSource
                isReplicated
                isArchived
                isReplica
                isExpired
            }
        }
    }
}`

// listSla GraphQL query
var listSlaQuery = `query SdkGolangListSla(
    $after: String,
    $filter: [GlobalSlaFilterInput!]) {
    globalSlaConnection(
        after: $after,
        filter: $filter,
    ) {
        edges {
            node {
                id,
                name,
                ... on GlobalSla {
                    baseFrequency {
                        duration,
                        unit,
                    },
                    objectTypes,
                    firstFullBackupWindows {
                        durationInHours,
                        startTimeAttributes {
                            dayOfWeek{
                                day,
                            },
                            hour,
                            minute,
                        }
                    },
                    backupWindows{
                        durationInHours,
                        startTimeAttributes {
                            dayOfWeek{
                                day,
                            },
                            hour,
                            minute,
                        }
                    },
                }
            }
        }
        pageInfo {
            endCursor,
            hasNextPage,
        },
        count,
    }
}`

// restoreK8sNamespace GraphQL query
var restoreK8sNamespaceQuery = `mutation SdkGolangRestoreK8sNamespace($k8sNamespaceRestoreRequest: K8sNamespaceRestore!) {
    restoreK8sNamespace(k8sNamespaceRestoreRequest: $k8sNamespaceRestoreRequest) {
        taskchainId
        jobId
    }
}`

// snapshotK8sNamespace GraphQL query
var snapshotK8sNamespaceQuery = `mutation SdkGolangSnapshotK8sNamespace($input: CreateK8sNamespaceSnapshotsInput!) {
    createK8sNamespaceSnapshots(input: $input) {
        taskchainId
        jobId
    }
}`
