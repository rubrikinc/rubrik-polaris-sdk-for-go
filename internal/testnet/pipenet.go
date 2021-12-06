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

package testnet

import (
	"context"
	"errors"
	"net"
	"net/http"
)

// pipeNetAddr dummy network address for a pipe network.
type pipeNetAddr struct{}

func (pipeNetAddr) Network() string {
	return "pipenet"
}

func (pipeNetAddr) String() string {
	return "pipenet"
}

// TestListener listens for incoming connection from the paired test dialer.
type TestListener struct {
	abort  chan struct{}
	accept chan net.Conn
}

func (l *TestListener) Accept() (net.Conn, error) {
	for {
		select {
		case <-l.abort:
			return nil, errors.New("closing test listener")
		case conn, ok := <-l.accept:
			if ok {
				return conn, nil
			}
		}
	}
}

func (l *TestListener) Close() error {
	close(l.abort)
	return nil
}

func (l *TestListener) Addr() net.Addr {
	return pipeNetAddr{}
}

// NewPipeNet returns a test dial/listener pair. The dial function returns a
// connection to the test listener which is backed by an in-memory pipe. Note
// that this function panics if http.DefaultTransport is not of type
// *http.Transport.
func NewPipeNet() (*http.Client, *TestListener) {
	accept := make(chan net.Conn)
	dialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
		client, server := net.Pipe()
		go func() {
			select {
			case accept <- server:
			case <-ctx.Done():
			}
		}()

		return client, nil
	}

	// Copy the default transport and install the test dialer.
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		panic("http.DefaultTransport is not of type *http.Transport")
	}
	transport = transport.Clone()
	transport.DialContext = dialer

	client := &http.Client{
		Transport: transport,
	}

	return client, &TestListener{abort: make(chan struct{}), accept: accept}
}
