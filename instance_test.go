package mockspanner_test

import (
	"context"
	"testing"

	mockspanner "github.com/arhea/go-mock-spanner"
)

func TestInstance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mock, err := mockspanner.NewInstance(ctx, t)

	if err != nil {
		t.Fatalf("creating the instance: %v", err)
		return
	}

	// close the mock
	defer mock.Close(ctx)

	port, err := mock.Port(ctx)

	if err != nil {
		t.Fatalf("getting the mapped port of the instance: %v", err)
		return
	}

	if port.Port() == "" {
		t.Fatalf("port should not be empty")
		return
	}
}
