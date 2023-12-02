# Mock Google Cloud Spanner

Provides a mock of [Google Cloud Spanner](https://cloud.google.com/spanner?hl=en) using the official [Google Cloud Spanner Emulator](https://github.com/GoogleCloudPlatform/cloud-spanner-emulator).

These mocks will automatically create a new emulator, wait for it to be available, then create a mock database. You will need to run your database migrations prior to performing your tests.

I recommend reusing the instance across multiple tests to reduce test run times.

This library is built on top of [testcontainers](https://testcontainers.com/).

## Usage

Creating a mock instance for creating a customer connection.

```golang
func TestXXX(t *testing.T) {
	ctx := context.Background()

	mock, err := mockspanner.NewInstance(ctx, t)

	if err != nil {
		t.Fatalf("creating the instance: %v", err)
		return
	}

	// close the mock
	defer mock.Close(ctx)

	// ... my test code
}
```

Creating a mock Spanner client for interacting with Spanner via the Go client.

```golang
func TestXXX(t *testing.T) {
	ctx := context.Background()

	mock, err := mockspanner.NewClient(ctx, t)

	if err != nil {
		t.Fatalf("creating the client: %v", err)
		return
	}

	// close the mock
	defer mock.Close(ctx)

    spannerClient := mock.Client()

    t.Run("MyTest1", func(t *testing.T) {
        // ... my test code
    })

    t.Run("MyTest2", func(t *testing.T) {
        // ... my test code
    })
}
```
