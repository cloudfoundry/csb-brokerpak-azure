// Package connector manages the creating and deletion of service bindings to MS SQL Server
package connector

import (
	"context"
	"fmt"
	sql2 "github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v3.0/sql"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/Azure/go-autorest/autorest/azure/auth"

)

func New(azureTenantID, azureClientID, azureClientSecret, azureSubscriptionID string) *Connector {
	return &Connector{
		azureTenantID:   azureTenantID,
		azureClientID: azureClientID,
		azureClientSecret: azureClientSecret,
		azureSubscriptionID: azureSubscriptionID,
	}
}

type Connector struct {
	azureTenantID, azureClientID, azureClientSecret, azureSubscriptionID string
}

func (c *Connector) CreateRunFailover(ctx context.Context, resourceGroup, serverName, failoverGroup string) error {
	return nil
}

func (c *Connector) DeleteRunFailover(ctx context.Context, resourceGroup, serverName, failoverGroup string) error {
	return nil
}

func (c *Connector) ReadRunFailover(ctx context.Context, resourceGroup, serverName, failoverGroup string) (result bool, err error) {
	return false, nil
}

// TODO: see if we can avoid setting environment variables
// TODO: sql2.NewFailoverGroupsClient dependency was forced in go.mod, see if can use a newer one.
func (c *Connector) withConnection(callback func(dbclient sql2.FailoverGroupsClient) error) error {
	os.Setenv("AZURE_SUBSCRIPTION_ID", c.azureSubscriptionID)
	os.Setenv("AZURE_TENANT_ID", c.azureTenantID)
	os.Setenv("AZURE_CLIENT_ID", c.azureClientID)
	os.Setenv("AZURE_CLIENT_SECRET", c.azureClientSecret)

	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return fmt.Errorf("error connecting: %w", err)
	}
	// Create AzureSQL SQL Failover Groups client
	dbclient := sql2.NewFailoverGroupsClient(c.azureSubscriptionID)
	dbclient.Authorizer = authorizer


	return callback(dbclient)
}
