package mockspanner

import (
	"context"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Instance represents the underlying container that is running the mock Spanner instance.
type Instance struct {
	t         *testing.T
	container testcontainers.Container
}

// Port returns the mapped port of the underlying container.
func (r *Instance) Port(ctx context.Context) (nat.Port, error) {
	r.t.Helper()

	return r.container.MappedPort(ctx, "9010")
}

// Close terminates the underlying container.
func (r *Instance) Close(ctx context.Context) {
	r.t.Helper()

	if err := r.container.Terminate(ctx); err != nil {
		r.t.Logf("error terminating spanner emulator: %v", err)
	}
}

func NewInstance(ctx context.Context, t *testing.T) (*Instance, error) {
	t.Helper()

	var err error

	// configure the backoff
	cfg := backoff.NewExponentialBackOff()
	cfg.InitialInterval = time.Second * 2
	cfg.MaxElapsedTime = time.Minute * 10
	policy := backoff.WithContext(cfg, ctx)

	// create the spanner emulator container
	operation := backoff.OperationWithData[testcontainers.Container](func() (testcontainers.Container, error) {
		return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				// https://github.com/GoogleCloudPlatform/cloud-spanner-emulator
				Image:        "gcr.io/cloud-spanner-emulator/emulator:latest",
				ExposedPorts: []string{"9010/tcp", "9020/tcp"},
				WaitingFor:   wait.ForLog("gRPC server listening at"),
			},
			Started: true,
			Reuse:   false,
			Logger:  testcontainers.TestLogger(t),
		})
	})

	// create the spanner emulator container
	spannerEmulator, err := backoff.RetryWithData(operation, policy)

	if err != nil {
		return nil, err
	}

	// create the mock instance
	cntr := &Instance{
		t:         t,
		container: spannerEmulator,
	}

	return cntr, nil
}
