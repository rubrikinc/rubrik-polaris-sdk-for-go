package cdm

import (
	"context"
	"testing"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testnet"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// assertCancel calls the cancel function and asserts it did not return an
// error.
func assertCancel(t *testing.T, cancel testnet.CancelFunc) {
	if err := cancel(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// newTestClient returns a new Client intended to be used by unit tests.
func newTestClient(logger log.Logger) (*Client, *testnet.TestListener) {
	testClient, listener := testnet.NewPipeNet()

	return &Client{
		client: &client{
			client: testClient,
		},
		nodeIP: "127.0.0.1",
		Log:    logger,
	}, listener
}
