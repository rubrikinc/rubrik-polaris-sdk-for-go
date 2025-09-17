// Copyright 2025 Rubrik, Inc.
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

// Package event provides high level functions when working with events GQL.

package event

import (
	gqlevent "github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/graphql/event"
)

// InProgress returns true if the event is in progress.
func InProgress(event gqlevent.EventSeries) bool {
	return event.LastActivityStatus == gqlevent.ActivityStatusQueued ||
		event.LastActivityStatus == gqlevent.ActivityStatusRunning ||
		event.LastActivityStatus == gqlevent.ActivityStatusTaskSuccess ||
		event.LastActivityStatus == gqlevent.ActivityStatusInfo ||
		event.LastActivityStatus == gqlevent.ActivityStatusWarning ||
		event.LastActivityStatus == gqlevent.ActivityStatusTaskFailure
}

// Success returns true if the event is successful.
func Success(event gqlevent.EventSeries) bool {
	return event.LastActivityStatus == gqlevent.ActivityStatusSuccess ||
		event.LastActivityStatus == gqlevent.ActivityStatusPartialSuccess
}

// Failure returns true if the event has failed.
func Failure(event gqlevent.EventSeries) bool {
	return event.LastActivityStatus == gqlevent.ActivityStatusFailure ||
		event.LastActivityStatus == gqlevent.ActivityStatusCanceled ||
		event.LastActivityStatus == gqlevent.ActivityStatusCanceling
}
