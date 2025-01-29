// Package testhelpers contains logic to create and retrieve data from the failover group using the Azure API
package testhelpers

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

type FailoverData struct {
	ResourceGroup              *armresources.ResourceGroup
	PartnerServerResourceGroup *armresources.ResourceGroup
	Server                     *armsql.Server
	PartnerServer              *armsql.Server
	FailoverGroup              *armsql.FailoverGroup
}

type FailoverConfig struct {
	ResourceGroupName, PartnerResourceGroupName, ServerName, MainLocation, PartnerServerName, SubscriptionID, FailoverGroupName, PartnerServerLocation string
}

func CreateFailoverGroup(cnf FailoverConfig) (FailoverData, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return FailoverData{}, err
	}
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred, cnf.ResourceGroupName, cnf.MainLocation, cnf.SubscriptionID)
	if err != nil {
		return FailoverData{}, err
	}

	server, err := createServer(ctx, cred, cnf.ResourceGroupName, cnf.ServerName, cnf.MainLocation, cnf.SubscriptionID)
	if err != nil {
		return FailoverData{}, err
	}

	partnerServer, err := createServer(ctx, cred, cnf.ResourceGroupName, cnf.PartnerServerName, cnf.PartnerServerLocation, cnf.SubscriptionID)
	if err != nil {
		return FailoverData{}, err
	}

	failoverGroup, err := createFailoverGroup(ctx, cred, *partnerServer.ID, cnf.ResourceGroupName, cnf.ServerName, cnf.FailoverGroupName, cnf.SubscriptionID)
	if err != nil {
		return FailoverData{}, err
	}

	return FailoverData{
		ResourceGroup:              resourceGroup,
		PartnerServerResourceGroup: resourceGroup,
		Server:                     server,
		PartnerServer:              partnerServer,
		FailoverGroup:              failoverGroup,
	}, nil
}

func Cleanup(cnf FailoverConfig) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return cleanup(ctx, cred, cnf.ResourceGroupName, cnf.SubscriptionID)
}

func GetFailoverGroup(resourceGroupName, serverName, failoverGroupName, subscriptionID string) (*armsql.FailoverGroup, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	return getFailoverGroup(ctx, cred, resourceGroupName, serverName, failoverGroupName, subscriptionID)
}

func createServer(ctx context.Context, cred azcore.TokenCredential, resourceGroupName, serverName, location, subscriptionID string) (*armsql.Server, error) {
	serversClient, err := armsql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		armsql.Server{
			Location: to.Ptr(location),
			Properties: &armsql.ServerProperties{
				AdministratorLogin:         to.Ptr(RandomName("dummylogin")),
				AdministratorLoginPassword: to.Ptr(RandomName("dummyPassword")),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Server, nil
}

func createFailoverGroup(ctx context.Context, cred azcore.TokenCredential, partnerServerID, resourceGroupName, serverName, failoverGroupName, subscriptionID string) (*armsql.FailoverGroup, error) {
	failoverGroupsClient, err := armsql.NewFailoverGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := failoverGroupsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		failoverGroupName,
		armsql.FailoverGroup{
			Properties: &armsql.FailoverGroupProperties{
				PartnerServers: []*armsql.PartnerInfo{
					{
						ID: to.Ptr(partnerServerID),
					},
				},
				ReadWriteEndpoint: &armsql.FailoverGroupReadWriteEndpoint{
					FailoverPolicy: to.Ptr(armsql.ReadWriteEndpointFailoverPolicyManual),
				},
				Databases: []*string{},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.FailoverGroup, nil
}

func getFailoverGroup(ctx context.Context, cred azcore.TokenCredential, resourceGroupName, serverName, failoverGroupName, subscriptionID string) (*armsql.FailoverGroup, error) {
	failoverGroupsClient, err := armsql.NewFailoverGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := failoverGroupsClient.Get(ctx, resourceGroupName, serverName, failoverGroupName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.FailoverGroup, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential, resourceGroupName, location, subscriptionID string) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential, resourceGroupName, subscriptionID string) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
