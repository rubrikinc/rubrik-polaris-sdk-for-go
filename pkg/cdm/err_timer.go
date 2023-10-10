// Copyright 2023 Rubrik, Inc.
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

package cdm

import "time"

type errTimer struct {
	timer   *time.Timer
	timeout time.Duration
	err     error
}

// newErrTimer returns a new stopped errTimer with the specified timeout.
func newErrTimer(timeout time.Duration) *errTimer {
	timer := time.NewTimer(10 * time.Second)
	if !timer.Stop() {
		<-timer.C
	}

	return &errTimer{timer: timer, timeout: timeout}
}

// expired returns true if the errTimer has expired, false otherwise.
func (t *errTimer) expired() <-chan time.Time {
	return t.timer.C
}

// reset records the error and starts the errTimer. If err is nil, the errTimer
// is stopped. If err is not nil and the timer is already running, the error is
// recorded and the timer is left unchanged.
func (t *errTimer) reset(err error) {
	switch {
	case err != nil && t.err == nil:
		t.timer.Reset(t.timeout)
	case err == nil && t.err != nil:
		if !t.timer.Stop() {
			<-t.timer.C
		}
	}

	t.err = err
}
