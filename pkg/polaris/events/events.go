package events

import (
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/events"
)

// IsEventInProgress returns true if the event is in progress.
func IsEventInProgress(event events.EventSeries) bool {
	return event.LastActivityStatus == events.ActivityStatusQueued ||
		event.LastActivityStatus == events.ActivityStatusRunning ||
		event.LastActivityStatus == events.ActivityStatusTaskSuccess ||
		event.LastActivityStatus == events.ActivityStatusInfo ||
		event.LastActivityStatus == events.ActivityStatusWarning ||
		event.LastActivityStatus == events.ActivityStatusTaskFailure
}

// IsEventSuccess returns true if the event is successful.
func IsEventSuccess(event events.EventSeries) bool {
	return event.LastActivityStatus == events.ActivityStatusSuccess ||
		event.LastActivityStatus == events.ActivityStatusPartialSuccess
}

// IsEventFailure returns true if the event has failed.
func IsEventFailure(event events.EventSeries) bool {
	return event.LastActivityStatus == events.ActivityStatusFailure ||
		event.LastActivityStatus == events.ActivityStatusCanceled ||
		event.LastActivityStatus == events.ActivityStatusCanceling
}
