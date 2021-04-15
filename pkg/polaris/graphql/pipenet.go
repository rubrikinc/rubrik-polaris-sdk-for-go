package graphql

import (
	"context"
	"errors"
	"net"
)

type testDialFunc func(context.Context, string, string) (net.Conn, error)

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
			return nil, errors.New("polaris: closing test listener")
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

// newPipeNet returns a test dial/listener pair. The dial function returns
// a connection to the test listener which is backed by an in-memory pipe.
func newPipeNet() (testDialFunc, *TestListener) {
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

	return dialer, &TestListener{abort: make(chan struct{}), accept: accept}
}
