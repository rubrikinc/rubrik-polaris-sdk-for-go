// Copyright 2026 Rubrik, Inc.
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

package event

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestEventSeriesFilterUsesInputEnumSpelling verifies that the activity status
// and type filters marshal to the EventStatus/EventType spelling RSC requires
// in filter input, and not to the mixed-case ActivityStatusEnum/ActivityTypeEnum
// spelling it returns in responses.
//
// RSC rejects the response spelling in a filter with an HTTP 400 during variable
// coercion, e.g. "Enum value 'Failure' is undefined in enum type 'EventStatus'".
func TestEventSeriesFilterUsesInputEnumSpelling(t *testing.T) {
	buf, err := json.Marshal(EventSeriesFilter{
		LastActivityStatus: []EventStatus{EventStatusFailure, EventStatusTaskFailure},
		LastActivityType:   []EventType{EventTypeBackup, EventTypeConfiguration},
		ObjectType:         []EventObjectType{EventObjectTypeCluster},
	})
	if err != nil {
		t.Fatal(err)
	}

	if got, want := string(buf), `{"lastActivityStatus":["FAILURE","TASK_FAILURE"],`+
		`"lastActivityType":["BACKUP","CONFIGURATION"],"objectType":["CLUSTER"]}`; got != want {
		t.Errorf("got filter JSON %s, want %s", got, want)
	}
}

// TestFilterEnumValuesAreScreamingSnakeCase verifies that every EventStatus and
// EventType constant uses the upper case, underscore separated spelling of the
// filter input enums. A value added in the response spelling would be silently
// rejected by RSC at query time rather than at compile time.
func TestFilterEnumValuesAreScreamingSnakeCase(t *testing.T) {
	values := make([]string, 0, len(allEventStatuses)+len(allEventTypes))
	for _, status := range allEventStatuses {
		values = append(values, string(status))
	}
	for _, eventType := range allEventTypes {
		values = append(values, string(eventType))
	}

	for _, value := range values {
		if value != strings.ToUpper(value) {
			t.Errorf("filter enum value %q is not upper case, it looks like the response spelling", value)
		}
	}
}

// TestResponseEnumsKeepResponseSpelling guards the other direction: the response
// enums deliberately keep RSC's mixed case spelling and must not be "corrected"
// to match the filter enums, or decoding an activity series breaks.
func TestResponseEnumsKeepResponseSpelling(t *testing.T) {
	if got, want := string(ActivityStatusFailure), "Failure"; got != want {
		t.Errorf("got ActivityStatusFailure %q, want %q", got, want)
	}
	if got, want := string(ActivityStatusTaskFailure), "TaskFailure"; got != want {
		t.Errorf("got ActivityStatusTaskFailure %q, want %q", got, want)
	}
	if got, want := string(ActivityTypeBackup), "Backup"; got != want {
		t.Errorf("got ActivityTypeBackup %q, want %q", got, want)
	}
}

var allEventStatuses = []EventStatus{
	EventStatusUnknown, EventStatusCanceled, EventStatusCanceling, EventStatusFailure,
	EventStatusInfo, EventStatusPartialSuccess, EventStatusQueued, EventStatusRunning,
	EventStatusSuccess, EventStatusTaskFailure, EventStatusTaskSuccess, EventStatusWarning,
}

var allEventTypes = []EventType{
	EventTypeUnknown, EventTypeAgentCloudSecurityAlert, EventTypeAnomaly, EventTypeArchive,
	EventTypeAuthDomain, EventTypeAwsEvent, EventTypeBackup, EventTypeBulkRecovery,
	EventTypeClassification, EventTypeCloudDirectArchive, EventTypeCloudNativeSource,
	EventTypeCloudNativeVirtualMachine, EventTypeCloudNativeVm, EventTypeConfiguration,
	EventTypeConnection, EventTypeConversion, EventTypeCopy, EventTypeDiagnostic,
	EventTypeDiscover, EventTypeDiscovery, EventTypeDownload, EventTypeEmbeddedEvent,
	EventTypeEncryptionManagementOperation, EventTypeFailover, EventTypeFileset,
	EventTypeHardware, EventTypeHdfs, EventTypeHostEvent, EventTypeHypervScvmm,
	EventTypeHypervServer, EventTypeIdentityActivity, EventTypeIdentityAlerts,
	EventTypeIdentityViolation, EventTypeIndex, EventTypeInstantiate, EventTypeIsolatedRecovery,
	EventTypeLegalHold, EventTypeLocalRecovery, EventTypeLockSnapshot, EventTypeLogBackup,
	EventTypeMaintenance, EventTypeNutanixCluster, EventTypeOwnership,
	EventTypePermissionAssessment, EventTypeProtectedObjectDeletion, EventTypeQuarantine,
	EventTypeRansomwareInvestigationAnalysis, EventTypeRecovery, EventTypeReencryption,
	EventTypeReplication, EventTypeResourceOperations, EventTypeScheduleRecovery,
	EventTypeSecurityViolation, EventTypeSeeding, EventTypeStorage, EventTypeStorageArray,
	EventTypeStormResource, EventTypeSupport, EventTypeSync, EventTypeSystem,
	EventTypeTenantOverlap, EventTypeTenantQuota, EventTypeTestFailover, EventTypeThreatFeed,
	EventTypeThreatHunt, EventTypeThreatMonitoring, EventTypeTpr, EventTypeUpgrade,
	EventTypeUserIntelligence, EventTypeVcd, EventTypeVcenter, EventTypeVolumeGroup,
}
