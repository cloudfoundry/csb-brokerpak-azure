// Package connector manages the creating and deletion of service bindings to MS SQL Server
package iaas

import (
	"context"
	"fmt"
	"os"

	"terraform-provider-csbmssqldbrunfailover/internal/failover"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

type Client struct {
	azureTenantID, azureClientID, azureClientSecret, azureSubscriptionID string
}

func NewClient(azureTenantID, azureClientID, azureClientSecret, azureSubscriptionID string) *Client {
	return &Client{
		azureTenantID:       azureTenantID,
		azureClientID:       azureClientID,
		azureClientSecret:   azureClientSecret,
		azureSubscriptionID: azureSubscriptionID,
	}
}

func (c *Client) CreateRunFailover(ctx context.Context, f failover.Failover) (string, error) {
	var failoverGroupID string
	partnerServer := &armsql.PartnerInfo{ID: to.Ptr(f.PartnerServerID())}

	readWriteEndpoint := &armsql.FailoverGroupReadWriteEndpoint{
		FailoverPolicy:                         to.Ptr(armsql.ReadWriteEndpointFailoverPolicyAutomatic),
		FailoverWithDataLossGracePeriodMinutes: to.Ptr[int32](int32(f.FailoverWithDataLossGracePeriodMinutes())),
	}

	err := c.withConnection(func(failoverGroupsClient *armsql.FailoverGroupsClient) error {

		pollerResp, err := failoverGroupsClient.BeginCreateOrUpdate(
			ctx,
			f.ResourceGroup(),
			f.ServerName(),
			f.FailoverGroup(),
			armsql.FailoverGroup{
				Properties: &armsql.FailoverGroupProperties{
					PartnerServers:    []*armsql.PartnerInfo{partnerServer},
					ReadWriteEndpoint: readWriteEndpoint,
					// TODO
					Databases:        []*string{},
					ReadOnlyEndpoint: nil,
					ReplicationRole:  nil,
					ReplicationState: nil,
				},
				Tags:     nil,
				ID:       nil,
				Location: nil,
				Name:     nil,
				Type:     nil,
			},
			nil,
		)
		if err != nil {
			return fmt.Errorf("error initiating the failover group creation operation %w", err)
		}

		resp, err := pollerResp.PollUntilDone(ctx, nil)
		if err != nil {
			return fmt.Errorf("error creating the failover group %w", err)
		}

		if resp.FailoverGroup.ID == nil {
			return fmt.Errorf("invalid fail over group id")
		}

		failoverGroupID = *resp.FailoverGroup.ID
		return nil
	})
	if err != nil {
		return "", err
	}

	return failoverGroupID, nil
}

func (c *Client) DeleteRunFailover(ctx context.Context, resourceGroup, serverName, failoverGroup string) error {
	return nil
}

func (c *Client) ReadRunFailover(ctx context.Context, resourceGroup, serverName, failoverGroup string) (result bool, err error) {
	return false, nil
}

func (c *Client) UpdateRunFailover(ctx context.Context, resourceGroup, serverName, failoverGroup string) (result bool, err error) {
	return false, nil
}

// TODO: see if we can avoid setting environment variables
// TODO: sql2.NewFailoverGroupsClient dependency was forced in go.mod, see if can use a newer one.
func (c *Client) withConnection(callback func(failoverGroupsClient *armsql.FailoverGroupsClient) error) error {
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
