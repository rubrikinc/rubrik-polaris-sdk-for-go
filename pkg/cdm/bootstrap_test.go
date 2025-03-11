package cdm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/internal/testnet"
	"github.com/rubrikinc/rubrik-polaris-sdk-for-go/pkg/polaris/log"
)

// TestIsBootstrapped verifies that the IsBootstrapped function parses and
// returns the correct value when the cluster returns a valid response.
func TestIsBootstrapped(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		if err := json.NewEncoder(w).Encode(struct {
			Value bool `json:"value"`
		}{Value: true}); err != nil {
			return err
		}
		return nil
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	ok, err := WrapBootstrap(client).IsBootstrapped(ctx, 1*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("invalid bootstrap status: false")
	}
}

// TestIsBootstrappedContextTimeout verifies that the IsBootstrapped function
// cancels its request and returns an error when the context is canceled.
func TestIsBootstrappedContextTimeout(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})
	defer assertCancel(t, cancel)

	ctx, ctxCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer ctxCancel()

	_, err := WrapBootstrap(client).IsBootstrapped(ctx, 1*time.Second, 100*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("invalid error type: %T", err)
	}
}

// TestIsBootstrappedErrorInvalid verifies that the IsBootstrapped function
// returns an error when the cluster returns an invalid response.
func TestIsBootstrappedErrorInvalid(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		return nil
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	_, err := WrapBootstrap(client).IsBootstrapped(ctx, 1*time.Second, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
}

// TestIsBootstrappedErrorRecover verifies that the IsBootstrapped function is
// able to recover from errors within the timeout duration.
func TestIsBootstrappedErrorRecover(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	count := 0
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		if count == 5 {
			if err := json.NewEncoder(w).Encode(struct {
				Value bool `json:"value"`
			}{Value: false}); err != nil {
				return err
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		count++
		return nil
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	ok, err := WrapBootstrap(client).IsBootstrapped(ctx, 1*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("invalid bootstrap status: true")
	}
}

// TestIsBootstrappedErrorTimeout verifies that the IsBootstrapped function
// returns an error if the error is not recovered within the timeout duration.
func TestIsBootstrappedErrorTimeout(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		w.WriteHeader(http.StatusNotFound)
		return nil
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	_, err := WrapBootstrap(client).IsBootstrapped(ctx, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "timeout waiting for bootstrap status:") {
		t.Fatalf("invalid error returned: %v", err)
	}
}

// TestBootstrap verifies that the BootstrapCluster function returns the values
// returned by the cluster in response to a bootstrap request.
func TestBootstrap(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		switch req.URL.Path {
		case "/api/internal/node_management/is_bootstrapped":
			if err := json.NewEncoder(w).Encode(struct {
				Value bool `json:"value"`
			}{Value: false}); err != nil {
				return err
			}
		case "/api/internal/cluster/me/bootstrap":
			w.WriteHeader(http.StatusAccepted)
			if err := json.NewEncoder(w).Encode(struct {
				ID int `json:"id"`
			}{ID: 1}); err != nil {
				return err
			}
		default:
			return errors.New("invalid url path")
		}
		return nil
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	id, err := WrapBootstrap(client).BootstrapCluster(ctx, ClusterConfig{}, 1*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatalf("invalid cluster id: %d", id)
	}
}

// TestBootstrapErrorIsBootstrapped verifies that the BootstrapCluster function
// returns the correct error when the IsBootstrapped function fails.
func TestBootstrapErrorIsBootstrapped(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		switch req.URL.Path {
		case "/api/internal/node_management/is_bootstrapped":
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		default:
			return errors.New("invalid url path")
		}
		return nil
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	_, err := WrapBootstrap(client).BootstrapCluster(ctx, ClusterConfig{}, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "failed to check cluster bootstrap status:") {
		t.Fatalf("invalid error returned: %v", err)
	}
}

// TestBootstrapErrorUnprocessableEntity verifies that the BootstrapCluster
// function returns the correct error when the bootstrap request returns an
// unprocessable entity error.
func TestBootstrapErrorUnprocessableEntity(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		switch req.URL.Path {
		case "/api/internal/node_management/is_bootstrapped":
			if err := json.NewEncoder(w).Encode(struct {
				Value bool `json:"value"`
			}{Value: false}); err != nil {
				return err
			}
		case "/api/internal/cluster/me/bootstrap":
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		default:
			return errors.New("invalid url path")
		}
		return nil
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	_, err := WrapBootstrap(client).BootstrapCluster(ctx, ClusterConfig{}, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "failed POST request") {
		t.Fatalf("invalid error returned: %v", err)
	}
}

// TestBootstrapErrorUnprocessableEntity verifies that the BootstrapCluster
// function returns the correct error when the bootstrap request returns an
// unprocessable entity error with a JSON body.
func TestBootstrapErrorUnprocessableEntityWithStatus(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		switch req.URL.Path {
		case "/api/internal/node_management/is_bootstrapped":
			if err := json.NewEncoder(w).Encode(struct {
				Value bool `json:"value"`
			}{Value: false}); err != nil {
				return err
			}
		case "/api/internal/cluster/me/bootstrap":
			w.WriteHeader(http.StatusUnprocessableEntity)
			if err := json.NewEncoder(w).Encode(struct {
				ID     int    `json:"id"`
				Status string `json:"status"`
			}{ID: 0, Status: "invalid configuration"}); err != nil {
				return err
			}
		default:
			return errors.New("invalid url path")
		}
		return nil
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	_, err := WrapBootstrap(client).BootstrapCluster(ctx, ClusterConfig{}, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "failed POST request") && !strings.HasSuffix(err.Error(), "invalid configuration") {
		t.Fatalf("invalid error returned: %v", err)
	}
}

// TestBootstrapWithBootstrappedCluster verifies that the BootstrapCluster
// function returns an error when the cluster is already bootstrapped.
func TestBootstrapWithBootstrappedCluster(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		switch req.URL.Path {
		case "/api/internal/node_management/is_bootstrapped":
			if err := json.NewEncoder(w).Encode(struct {
				Value bool `json:"value"`
			}{Value: true}); err != nil {
				return err
			}
		default:
			return errors.New("invalid url path")
		}
		return nil
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	_, err := WrapBootstrap(client).BootstrapCluster(ctx, ClusterConfig{}, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "cluster is already bootstrapped") {
		t.Fatalf("invalid error returned: %v", err)
	}
}

// TestWaitForBootstrap verifies that the WaitForBootstrap function parses and
// returns the correct value when the cluster returns a valid response to the
// bootstrap GET request.
func TestWaitForBootstrap(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		switch req.URL.Path {
		case "/api/internal/cluster/me/bootstrap":
			if err := json.NewEncoder(w).Encode(struct {
				Status  string `json:"status"`
				Message string `json:"message"`
			}{Status: "", Message: ""}); err != nil {
				return err
			}
			return nil
		default:
			return errors.New("invalid url path")
		}
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	err := WrapBootstrap(client).WaitForBootstrap(ctx, 1, 1*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
}

// TestWaitForBootstrapContextTimeout verifies that the WaitForBootstrap
// function cancels its request and returns an error when the context is
// canceled.
func TestWaitForBootstrapContextTimeout(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		switch req.URL.Path {
		case "/api/internal/cluster/me/bootstrap":
			if err := json.NewEncoder(w).Encode(struct {
				Status  string `json:"status"`
				Message string `json:"message"`
			}{Status: "IN_PROGRESS", Message: "bootstrap in progress"}); err != nil {
				return err
			}
			return nil
		default:
			return errors.New("invalid url path")
		}
	})
	defer assertCancel(t, cancel)

	ctx, ctxCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer ctxCancel()

	err := WrapBootstrap(client).WaitForBootstrap(ctx, 1, 1*time.Second, 100*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("invalid error type: %T", err)
	}
}

// TestWaitForBootstrapErrorRecover verifies that the WaitForBootstrap function
// is able to recover from errors within the timeout duration.
func TestWaitForBootstrapErrorRecover(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	count := 0
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		switch req.URL.Path {
		case "/api/internal/cluster/me/bootstrap":
			if count == 5 {
				if err := json.NewEncoder(w).Encode(struct {
					Status  string `json:"status"`
					Message string `json:"message"`
				}{Status: "", Message: "hshshshs"}); err != nil {
					return err
				}
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			count++
			return nil
		default:
			return errors.New("invalid url path")
		}
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	err := WrapBootstrap(client).WaitForBootstrap(ctx, 1, 1*time.Second, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
}

// TestWaitForBootstrapErrorTimeout verifies that the WaitForBootstrap function
// returns an error if the error is not recovered within the timeout duration.
func TestWaitForBootstrapErrorTimeout(t *testing.T) {
	client, lis := newTestClient(log.DiscardLogger{})
	cancel := testnet.ServeJSONWithTLS(lis, func(w http.ResponseWriter, req *http.Request) error {
		switch req.URL.Path {
		case "/api/internal/cluster/me/bootstrap":
			w.WriteHeader(http.StatusServiceUnavailable)
			return nil
		default:
			return errors.New("invalid url path")
		}
	})
	defer assertCancel(t, cancel)

	ctx := context.Background()
	err := WrapBootstrap(client).WaitForBootstrap(ctx, 1, 1*time.Second, 100*time.Millisecond)
	if err == nil || !strings.HasPrefix(err.Error(), "timeout waiting for bootstrap:") {
		t.Fatalf("invalid error returned: %v", err)
	}
}
