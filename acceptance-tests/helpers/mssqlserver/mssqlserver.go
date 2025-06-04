// Package mssqlserver manages database servers and failover groups configuration
package mssqlserver

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"csbbrokerpakazure/acceptance-tests/helpers/environment"
	"csbbrokerpakazure/acceptance-tests/helpers/random"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
)

const mainLocation = "westus2"

// DatabaseServerPairConfig represents a pair of database servers
type DatabaseServerPairConfig struct {
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

// NewDatabaseServerPairConfig creates a new database server pair configuration
func NewDatabaseServerPairConfig(metadata environment.Metadata) DatabaseServerPairConfig {
	primaryResourceGroup := random.Name(random.WithPrefix(metadata.ResourceGroup))
	secondaryResourceGroup := random.Name(random.WithPrefix(metadata.ResourceGroup))

	return DatabaseServerPairConfig{
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
func (d DatabaseServerPairConfig) PrimaryConfig() any {
	return d.memberConfig(d.PrimaryServer.Name, mainLocation, d.PrimaryServer.ResourceGroup)
}

// SecondaryConfig returns the configuration for the secondary database server
func (d DatabaseServerPairConfig) SecondaryConfig() any {
	return d.memberConfig(d.SecondaryServer.Name, mainLocation, d.SecondaryServer.ResourceGroup)
}

// memberConfig returns the configuration for a database server
func (d DatabaseServerPairConfig) memberConfig(name, location, rg string) any {
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
func (d DatabaseServerPairConfig) SecondaryResourceGroupConfig() any {
	return struct {
		InstanceName string `json:"instance_name"`
		Location     string `json:"location"`
	}{
		InstanceName: d.SecondaryResourceGroup,
		Location:     mainLocation,
	}
}

func (d DatabaseServerPairConfig) ServerPairsConfig() any {
	return map[string]any{d.ServerPairTag: d}
}

// CreateServerPair creates a new database server pair
func CreateServerPair(metadata environment.Metadata, firewallStartIP, firewallEndIP string, subscriptionID string) DatabaseServerPairConfig {
	cnf := NewDatabaseServerPairConfig(metadata)

	CreateResourceGroup(cnf.PrimaryServer.ResourceGroup, subscriptionID)
	CreateServer(cnf.PrimaryServer, cnf.Username, cnf.Password, subscriptionID)
	CreateFirewallRule(metadata, firewallStartIP, firewallEndIP, cnf.PrimaryServer, subscriptionID)
	CreateResourceGroup(cnf.SecondaryServer.ResourceGroup, subscriptionID)
	CreateServer(cnf.SecondaryServer, cnf.Username, cnf.Password, subscriptionID)
	CreateFirewallRule(metadata, firewallStartIP, firewallEndIP, cnf.SecondaryServer, subscriptionID)

	return cnf
}

func CreateFirewallRule(metadata environment.Metadata, firewallStartIP, firewallEndIP string, member DatabaseServer, subscriptionID string) {
	cred := must(azidentity.NewDefaultAzureCredential(nil))
	firewallClient := must(armsql.NewFirewallRulesClient(subscriptionID, cred, nil))

	// Use PublicIP from metadata if no overrides were specified
	if firewallStartIP == "" && firewallEndIP == "" && metadata.PublicIP != "" {
		GinkgoWriter.Println("Using public IP from metadata")
		firewallStartIP = metadata.PublicIP
		firewallEndIP = metadata.PublicIP
	}

	// Skip firewall rule creation if there are no IPs available
	if firewallStartIP == "" || firewallEndIP == "" {
		GinkgoWriter.Println("Skipping firewall rule creation")
		return
	}

	_, err := firewallClient.CreateOrUpdate(
		context.Background(),
		member.ResourceGroup,
		member.Name,
		"firewallrule-"+member.Name,
		armsql.FirewallRule{
			Properties: &armsql.ServerFirewallRuleProperties{
				StartIPAddress: to.Ptr(firewallStartIP),
				EndIPAddress:   to.Ptr(firewallEndIP),
			},
		},
		nil,
	)
	Expect(err).NotTo(HaveOccurred())
}

func CleanupFirewallRule(member DatabaseServer, subscriptionID string) {
	cred := must(azidentity.NewDefaultAzureCredential(nil))
	firewallClient := must(armsql.NewFirewallRulesClient(subscriptionID, cred, nil))

	_, err := firewallClient.Delete(context.Background(), member.ResourceGroup, member.Name, "firewallrule-"+member.Name, nil)
	Expect(err).NotTo(HaveOccurred())
}

func CleanupServer(member DatabaseServer, subscriptionID string) {
	cred := must(azidentity.NewDefaultAzureCredential(nil))
	serversClient := must(armsql.NewServersClient(subscriptionID, cred, nil))

	pollerResp := must(serversClient.BeginDelete(context.Background(), member.ResourceGroup, member.Name, nil))

	_, err := pollerResp.PollUntilDone(context.Background(), nil)
	Expect(err).NotTo(HaveOccurred())
}

func Cleanup(cnf DatabaseServerPairConfig, subscriptionID string) {
	CleanupFirewallRule(cnf.PrimaryServer, subscriptionID)
	CleanupServer(cnf.PrimaryServer, subscriptionID)
	cleanupResourceGroup(cnf.ResourceGroup, subscriptionID)
	CleanupFirewallRule(cnf.SecondaryServer, subscriptionID)
	CleanupServer(cnf.SecondaryServer, subscriptionID)
	cleanupResourceGroup(cnf.SecondaryResourceGroup, subscriptionID)
}

func cleanupResourceGroup(resourceGroupName, subscriptionID string) {
	cred := must(azidentity.NewDefaultAzureCredential(nil))
	resourceGroupClient := must(armresources.NewResourceGroupsClient(subscriptionID, cred, nil))

	pollerResp := must(resourceGroupClient.BeginDelete(context.Background(), resourceGroupName, nil))

	_, err := pollerResp.PollUntilDone(context.Background(), nil)
	Expect(err).NotTo(HaveOccurred())
}

func CreateServer(member DatabaseServer, username, password, subscriptionID string) {
	cred := must(azidentity.NewDefaultAzureCredential(nil))
	serversClient := must(armsql.NewServersClient(subscriptionID, cred, nil))

	pollerResp := must(serversClient.BeginCreateOrUpdate(
		context.Background(),
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
	))

	_, err := pollerResp.PollUntilDone(context.Background(), nil)
	Expect(err).NotTo(HaveOccurred())
}

func CreateResourceGroup(resourceGroupName, subscriptionID string) {
	cred := must(azidentity.NewDefaultAzureCredential(nil))
	resourceGroupClient := must(armresources.NewResourceGroupsClient(subscriptionID, cred, nil))

	_, err := resourceGroupClient.CreateOrUpdate(context.Background(), resourceGroupName, armresources.ResourceGroup{Location: to.Ptr(mainLocation)}, nil)
	Expect(err).NotTo(HaveOccurred())
}

func must[A any](input A, err error) A {
	GinkgoHelper()

	Expect(err).NotTo(HaveOccurred())
	return input
}
