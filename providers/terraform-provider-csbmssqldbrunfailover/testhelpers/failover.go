package testhelpers

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/onsi/ginkgo/v2"
)

type FailoverData struct {
	ResourceGroup *armresources.ResourceGroup
	Server        *armsql.Server
	PartnerServer *armsql.Server
	FailoverGroup *armsql.FailoverGroup
}

type FailoverConfig struct {
	ResourceGroupName, ServerName, Location, PartnerServerName, SubscriptionID, FailoverGroupName string
}

func CreateFailoverGroup(cnf FailoverConfig) FailoverData {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred, cnf.ResourceGroupName, cnf.Location, cnf.SubscriptionID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	server, err := createServer(ctx, cred, cnf.ResourceGroupName, cnf.ServerName, cnf.Location, cnf.SubscriptionID)
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("error creating server %s", err))
	}

	partnerServer, err := createPartnerServer(ctx, cred, cnf.ResourceGroupName, cnf.PartnerServerName, cnf.SubscriptionID)
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("error creating partner server %s", err))
	}

	failoverGroup, err := createFailoverGroup(ctx, cred, *partnerServer.ID, cnf.ResourceGroupName, cnf.ServerName, cnf.FailoverGroupName, cnf.SubscriptionID)
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("error creating failover %s", err))
	}
	log.Println("failover group:", *failoverGroup.ID)

	return FailoverData{
		ResourceGroup: resourceGroup,
		Server:        server,
		PartnerServer: partnerServer,
		FailoverGroup: failoverGroup,
	}
}

func GetFailoverGroup(resourceGroupName, serverName, failoverGroupName, subscriptionID string) (*armsql.FailoverGroup, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
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
				AdministratorLogin:         to.Ptr("dummylogin"),
				AdministratorLoginPassword: to.Ptr("QWE123!@#"),
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

func createPartnerServer(ctx context.Context, cred azcore.TokenCredential, resourceGroupName, partnerServerName, subscriptionID string) (*armsql.Server, error) {
	serversClient, err := armsql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		partnerServerName,
		armsql.Server{
			Location: to.Ptr("eastus2"),
			Properties: &armsql.ServerProperties{
				AdministratorLogin:         to.Ptr("dummylogin"),
				AdministratorLoginPassword: to.Ptr("QWE123!@#"),
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
					FailoverPolicy:                         to.Ptr(armsql.ReadWriteEndpointFailoverPolicyAutomatic),
					FailoverWithDataLossGracePeriodMinutes: to.Ptr[int32](480),
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

func Cleanup(ctx context.Context, cred azcore.TokenCredential, resourceGroupName, subscriptionID string) error {
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
