package mockspanner

import (
	"context"
	"fmt"
	"os"
	"testing"

	"cloud.google.com/go/spanner"
	databaseAdmin "cloud.google.com/go/spanner/admin/database/apiv1"
	adminpb "cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instanceAdmin "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client represents the underlying container that is running the mock Spanner instance. It includes all of the
// clients that are needed to interact with the instance, database, and user client.
type Client struct {
	t                   *testing.T
	instance            *Instance
	grpc                *grpc.ClientConn
	client              *spanner.Client
	dbAdminClient       *databaseAdmin.DatabaseAdminClient
	instanceAdminClient *instanceAdmin.InstanceAdminClient
}

// GRPC fetches the connection to the underlying Spanner gRPC server.
func (r *Client) GRPC() *grpc.ClientConn {
	r.t.Helper()

	return r.grpc
}

// Client returns the underlying Spanner client.
func (r *Client) Client() *spanner.Client {
	r.t.Helper()

	return r.client
}

// Client returns the underlying Spanner Instance Admin client.
func (r *Client) InstanceAdmin() *instanceAdmin.InstanceAdminClient {
	r.t.Helper()

	return r.instanceAdminClient
}

// Client returns the underlying Spanner Database Admin client.
func (r *Client) DatabaseAdmin() *databaseAdmin.DatabaseAdminClient {
	r.t.Helper()

	return r.dbAdminClient
}

// Close closes the underlying Spanner clients and the instance.
func (r *Client) Close(ctx context.Context) {
	r.t.Helper()

	r.client.Close()

	if err := r.instanceAdminClient.Close(); err != nil {
		r.t.Logf("error closing instance admin client: %v", err)
	}

	if err := r.dbAdminClient.Close(); err != nil {
		r.t.Logf("error closing database admin client: %v", err)
	}

	r.instance.Close(ctx)
}

func NewClient(ctx context.Context, t *testing.T) (*Client, error) {
	t.Helper()

	instance, err := NewInstance(ctx, t)

	if err != nil {
		return nil, fmt.Errorf("creating the instance: %v", err)
	}

	spannerPort, err := instance.Port(ctx)

	if err != nil {
		return nil, fmt.Errorf("getting the mapped port of the instance: %v", err)
	}

	// nolint:gosec
	if err := os.Setenv("GCLOUD_PROJECT", ProjectID); err != nil {
		return nil, fmt.Errorf("setting GCLOUD_PROJECT environment variable: %v", err)
	}

	// nolint:gosec
	if err := os.Setenv("SPANNER_EMULATOR_HOST", "localhost:"+spannerPort.Port()); err != nil {
		return nil, fmt.Errorf("setting SPANNER_EMULATOR_HOST environment variable: %v", err)
	}

	fullDatabaseName := fmt.Sprintf("projects/%s/instances/%s/databases/%s", ProjectID, InstanceID, DatabaseID)

	conn, err := grpc.Dial("localhost:"+spannerPort.Port(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	// create the spanner mock clients
	instanceAdminClient, err := instanceAdmin.NewInstanceAdminClient(ctx, option.WithGRPCConn(conn))

	if err != nil {
		return nil, fmt.Errorf("creating instance admin client: %v", err)
	}

	_, err = instanceAdminClient.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: "projects/" + ProjectID + "/instances/" + InstanceID,
	})

	if err != nil {

		_, err = instanceAdminClient.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
			Parent:     "projects/" + ProjectID,
			InstanceId: InstanceID,
		})

		if err != nil {
			return nil, fmt.Errorf("creating instance: %v", err)
		}

	}

	adminClient, err := databaseAdmin.NewDatabaseAdminClient(ctx, option.WithGRPCConn(conn))

	if err != nil {
		return nil, fmt.Errorf("creating database admin client: %v", err)
	}

	op, err := adminClient.CreateDatabase(ctx, &adminpb.CreateDatabaseRequest{
		Parent:          "projects/" + ProjectID + "/instances/" + InstanceID,
		CreateStatement: "CREATE DATABASE `" + DatabaseID + "`",
	})

	if err != nil {
		return nil, fmt.Errorf("creating database: %v", err)
	}

	_, err = op.Wait(ctx)

	if err != nil {
		return nil, fmt.Errorf("waiting for database to be created: %v", err)
	}

	spannerClient, err := spanner.NewClient(ctx, fullDatabaseName, option.WithGRPCConn(conn), option.WithGRPCConnectionPool(10))

	if err != nil {
		return nil, fmt.Errorf("creating spanner client: %v", err)
	}

	client := &Client{
		t:                   t,
		instance:            instance,
		grpc:                conn,
		client:              spannerClient,
		dbAdminClient:       adminClient,
		instanceAdminClient: instanceAdminClient,
	}

	return client, nil
}
