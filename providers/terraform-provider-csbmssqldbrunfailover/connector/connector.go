// Package connector manages the failover state in MSSQL DB
package connector

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

type Connector struct {
	azureTenantID, azureClientID, azureClientSecret, azureSubscriptionID string
}

func NewConnector(azureTenantID, azureClientID, azureClientSecret, azureSubscriptionID string) *Connector {
	return &Connector{
		azureTenantID:       azureTenantID,
		azureClientID:       azureClientID,
		azureClientSecret:   azureClientSecret,
		azureSubscriptionID: azureSubscriptionID,
	}
}

func (c *Connector) RunFailover(ctx context.Context, resourceGroup, serverName, failoverGroup string) error {

	return c.withConnection(func(failoverGroupsClient *armsql.FailoverGroupsClient) error {
		pollerResp, err := failoverGroupsClient.BeginFailover(ctx, resourceGroup, serverName, failoverGroup, nil)
		if err != nil {
			return fmt.Errorf("error initiating failover action %w", err)
		}

		_, err = pollerResp.PollUntilDone(ctx, nil)
		if err != nil {
			return fmt.Errorf("error activating failover %w", err)
		}

		return nil
	})
}

func (c *Connector) ReadRunFailover(ctx context.Context, resourceGroup, serverName, failoverGroup string) (bool, error) {
	var existFailover bool
	err := c.withConnection(func(failoverGroupsClient *armsql.FailoverGroupsClient) error {
		f, err := failoverGroupsClient.Get(ctx, resourceGroup, serverName, failoverGroup, nil)
		if err != nil {
			return fmt.Errorf("error getting failover %w", err)
		}

		existFailover = *f.Name == failoverGroup

		return nil
	})
	if err != nil {
		return false, err
	}

	return existFailover, nil
}

func (c *Connector) withConnection(callback func(failoverGroupsClient *armsql.FailoverGroupsClient) error) error {
	_ = os.Setenv("AZURE_SUBSCRIPTION_ID", c.azureSubscriptionID)
	_ = os.Setenv("AZURE_TENANT_ID", c.azureTenantID)
	_ = os.Setenv("AZURE_CLIENT_ID", c.azureClientID)
	_ = os.Setenv("AZURE_CLIENT_SECRET", c.azureClientSecret)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("azure credential could not be created %w", err)
	}

	failoverGroupsClient, err := armsql.NewFailoverGroupsClient(c.azureSubscriptionID, cred, nil)
	if err != nil {
		return fmt.Errorf("azure failover groups client could not be created %w", err)
	}

	return callback(failoverGroupsClient)
}
