// Package mssqlserver manages database servers and failover groups configuration
package mssqlserver

import (
	"context"

	"csbbrokerpakazure/acceptance-tests/helpers/environment"
	"csbbrokerpakazure/acceptance-tests/helpers/random"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

const (
	mainLocation = "westus2"
)

// DatabaseServerPairCnf represents a pair of database servers
type DatabaseServerPairCnf struct {
	ServerPairTag          string
	Username               string         `json:"admin_username"`
	Password               string         `json:"admin_password"`
	PrimaryServer          DatabaseServer `json:"primary"`
	SecondaryServer        DatabaseServer `json:"secondary"`
	SecondaryResourceGroup string         `json:"-"`
	ResourceGroup          string         `json:"-"`
}

// DatabaseServer represents a database server
type DatabaseServer struct {
	Name          string `json:"server_name"`
	ResourceGroup string `json:"resource_group"`
}

// NewDatabaseServerPairCnf creates a new database server pair configuration
func NewDatabaseServerPairCnf(metadata environment.Metadata) DatabaseServerPairCnf {
	primaryResourceGroup := random.Name(random.WithPrefix(metadata.ResourceGroup))
	secondaryResourceGroup := random.Name(random.WithPrefix(metadata.ResourceGroup))

	return DatabaseServerPairCnf{
		ServerPairTag: random.Name(random.WithMaxLength(10)),
		Username:      random.Name(random.WithMaxLength(10)),
		Password:      random.Password(),
		PrimaryServer: DatabaseServer{
			Name:          random.Name(random.WithPrefix("server")),
			ResourceGroup: primaryResourceGroup,
		},
		SecondaryServer: DatabaseServer{
			Name:          random.Name(random.WithPrefix("server")),
			ResourceGroup: secondaryResourceGroup,
		},
		ResourceGroup:          primaryResourceGroup,
		SecondaryResourceGroup: secondaryResourceGroup,
	}
}

// PrimaryConfig returns the configuration for the primary database server
func (d DatabaseServerPairCnf) PrimaryConfig() any {
	return d.memberConfig(d.PrimaryServer.Name, mainLocation, d.PrimaryServer.ResourceGroup)
}

// SecondaryConfig returns the configuration for the secondary database server
func (d DatabaseServerPairCnf) SecondaryConfig() any {
	return d.memberConfig(d.SecondaryServer.Name, mainLocation, d.SecondaryServer.ResourceGroup)
}

// memberConfig returns the configuration for a database server
func (d DatabaseServerPairCnf) memberConfig(name, location, rg string) any {
	return struct {
		Name          string `json:"instance_name"`
		Username      string `json:"admin_username"`
		Password      string `json:"admin_password"`
		Location      string `json:"location"`
		ResourceGroup string `json:"resource_group"`
	}{
		Name:          name,
		Username:      d.Username,
		Password:      d.Password,
		Location:      location,
		ResourceGroup: rg,
	}
}

// SecondaryResourceGroupConfig returns the configuration for the secondary resource group
func (d DatabaseServerPairCnf) SecondaryResourceGroupConfig() any {
	return struct {
		InstanceName string `json:"instance_name"`
		Location     string `json:"location"`
	}{
		InstanceName: d.SecondaryResourceGroup,
		Location:     mainLocation,
	}
}

func (d DatabaseServerPairCnf) ServerPairsConfig() any {
	return map[string]any{d.ServerPairTag: d}
}

// CreateServerPair creates a new database server pair
func CreateServerPair(ctx context.Context, metadata environment.Metadata, subscriptionID string) (DatabaseServerPairCnf, error) {

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return DatabaseServerPairCnf{}, err
	}

	cnf := NewDatabaseServerPairCnf(metadata)

	if err := createResourceGroup(ctx, cred, cnf.PrimaryServer.ResourceGroup, subscriptionID); err != nil {
		return DatabaseServerPairCnf{}, err
	}

	if err := createServer(ctx, cred, cnf.PrimaryServer, cnf.Username, cnf.Password, subscriptionID); err != nil {
		return DatabaseServerPairCnf{}, err
	}

	if err := createFirewallRule(ctx, cred, metadata, cnf.PrimaryServer, subscriptionID); err != nil {
		return DatabaseServerPairCnf{}, err
	}

	if err := createResourceGroup(ctx, cred, cnf.SecondaryServer.ResourceGroup, subscriptionID); err != nil {
		return DatabaseServerPairCnf{}, err
	}

	if err := createServer(ctx, cred, cnf.SecondaryServer, cnf.Username, cnf.Password, subscriptionID); err != nil {
		return DatabaseServerPairCnf{}, err
	}

	if err := createFirewallRule(ctx, cred, metadata, cnf.SecondaryServer, subscriptionID); err != nil {
		return DatabaseServerPairCnf{}, err
	}

	return cnf, nil
}

func CreateFirewallRule(ctx context.Context, metadata environment.Metadata, member DatabaseServer, subscriptionID string) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	return createFirewallRule(ctx, cred, metadata, member, subscriptionID)
}

func createFirewallRule(ctx context.Context, cred azcore.TokenCredential, metadata environment.Metadata, member DatabaseServer, subscriptionID string) error {
	firewallClient, err := armsql.NewFirewallRulesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	_, err = firewallClient.CreateOrUpdate(
		ctx,
		member.ResourceGroup,
		member.Name,
		"firewallrule-"+member.Name,
		armsql.FirewallRule{
			Properties: &armsql.ServerFirewallRuleProperties{
				StartIPAddress: to.Ptr(metadata.PublicIP),
				EndIPAddress:   to.Ptr(metadata.PublicIP),
			},
		},
		nil,
	)
	if err != nil {
		return err
	}

	return nil

}

func CleanFirewallRule(ctx context.Context, member DatabaseServer, subscriptionID string) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	return cleanupFirewallRule(ctx, cred, member, subscriptionID)
}

func cleanupFirewallRule(ctx context.Context, cred azcore.TokenCredential, member DatabaseServer, subscriptionID string) error {
	firewallClient, err := armsql.NewFirewallRulesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	_, err = firewallClient.Delete(ctx, member.ResourceGroup, member.Name, "firewallrule-"+member.Name, nil)
	if err != nil {
		return err
	}

	return nil
}

func CleanupServer(ctx context.Context, member DatabaseServer, subscriptionID string) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	return cleanupServer(ctx, cred, member, subscriptionID)
}

func cleanupServer(ctx context.Context, cred azcore.TokenCredential, member DatabaseServer, subscriptionID string) error {

	serversClient, err := armsql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := serversClient.BeginDelete(ctx, member.ResourceGroup, member.Name, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func Cleanup(ctx context.Context, cnf DatabaseServerPairCnf, subscriptionID string) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	if err := cleanupFirewallRule(ctx, cred, cnf.PrimaryServer, subscriptionID); err != nil {
		return err
	}

	if err := cleanupServer(ctx, cred, cnf.PrimaryServer, subscriptionID); err != nil {
		return err
	}

	if err := cleanupResourceGroup(ctx, cred, cnf.ResourceGroup, subscriptionID); err != nil {
		return err
	}

	if err := cleanupFirewallRule(ctx, cred, cnf.SecondaryServer, subscriptionID); err != nil {
		return err
	}

	if err := cleanupServer(ctx, cred, cnf.SecondaryServer, subscriptionID); err != nil {
		return err
	}

	if err := cleanupResourceGroup(ctx, cred, cnf.SecondaryResourceGroup, subscriptionID); err != nil {
		return err
	}

	return nil
}

func cleanupResourceGroup(ctx context.Context, cred azcore.TokenCredential, resourceGroupName, subscriptionID string) error {
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

func CreateServer(ctx context.Context, member DatabaseServer, username, password, subscriptionID string) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	return createServer(ctx, cred, member, username, password, subscriptionID)
}

func createServer(ctx context.Context, cred azcore.TokenCredential, member DatabaseServer, username, password, subscriptionID string) error {
	serversClient, err := armsql.NewServersClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := serversClient.BeginCreateOrUpdate(
		ctx,
		member.ResourceGroup,
		member.Name,
		armsql.Server{
			Location: to.Ptr(mainLocation),
			Properties: &armsql.ServerProperties{
				AdministratorLogin:         to.Ptr(username),
				AdministratorLoginPassword: to.Ptr(password),
			},
		},
		nil,
	)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential, resourceGroupName, subscriptionID string) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	_, err = resourceGroupClient.CreateOrUpdate(ctx, resourceGroupName, armresources.ResourceGroup{Location: to.Ptr(mainLocation)}, nil)
	if err != nil {
		return err
	}

	return nil
}

func CreateResourceGroup(ctx context.Context, resourceGroupName, subscriptionID string) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	return createResourceGroup(ctx, cred, resourceGroupName, subscriptionID)
}
