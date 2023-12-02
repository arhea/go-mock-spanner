package mockspanner_test

import (
	"context"
	"testing"

	mockspanner "github.com/arhea/go-mock-spanner"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mock, err := mockspanner.NewClient(ctx, t)

	if err != nil {
		t.Fatalf("creating the client: %v", err)
		return
	}

	// close the mock
	defer mock.Close(ctx)

	assert.NotNil(t, mock.Client(), "client should not be nil")
}
