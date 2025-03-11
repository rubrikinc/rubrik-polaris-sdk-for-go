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

package testnet

import (
	"context"
	"net"
	"net/http"
	"sync"
)

// CancelFunc stops the serving of the HandlerFunc function and returns any
// error encountered while serving the function.
type CancelFunc func(ctx context.Context) error

// HandlerFunc is a function that handles HTTP/HTTPS requests.
// Any error returned from the function will be returned by the CancelFunc
// function.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// serverFunc is a function that starts a server serving the HandlerFunc
// function until canceled by the CancelFunc.
type serverFunc func(lis net.Listener, server *http.Server) error

// serve starts a server, using the serverFunc, serving the HandlerFunc function
// until canceled by the CancelFunc.
func serve(lis net.Listener, server serverFunc, handler HandlerFunc) CancelFunc {
	ch := make(chan error)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Wait()
		close(ch)
	}()

	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			rb := newResponseBuffer(w)
			if err := handler(rb, req); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				wg.Add(1)
				go func() {
					defer wg.Done()
					ch <- err
				}()
				return
			}
			rb.copyTo(w)
		}),
	}

	go func() {
		defer wg.Done()
		if err := server(lis, srv); err != nil {
			ch <- err
		}
	}()

	return func(ctx context.Context) error {
		err := srv.Shutdown(ctx)

		// Flush the channel, keeping the first error encountered.
		for e := range ch {
			if err == nil {
				err = e
			}
		}

		return err
	}
}
