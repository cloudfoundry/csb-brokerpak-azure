package mssql_helpers

import "acceptancetests/helpers"

type DatabaseServerPair struct {
	ServerPairTag          string
	Username               string                   `json:"admin_username"`
	Password               string                   `json:"admin_password"`
	PrimaryServer          DatabaseServerPairMember `json:"primary"`
	SecondaryServer        DatabaseServerPairMember `json:"secondary"`
	SecondaryResourceGroup string                   `json:"-"`
}

type DatabaseServerPairMember struct {
	Name          string `json:"server_name"`
	ResourceGroup string `json:"resource_group"`
}

func (d DatabaseServerPair) PrimaryConfig() interface{} {
	return d.memberConfig(d.PrimaryServer.Name, "westus", d.PrimaryServer.ResourceGroup)
}

func (d DatabaseServerPair) SecondaryConfig() interface{} {
	return d.memberConfig(d.SecondaryServer.Name, "eastus", d.SecondaryServer.ResourceGroup)
}

func (d DatabaseServerPair) memberConfig(name, location, rg string) interface{} {
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

func (d DatabaseServerPair) SecondaryResourceGroupConfig() interface{} {
	return struct {
		InstanceName string `json:"instance_name"`
		Location     string `json:"location"`
	}{
		InstanceName: d.SecondaryResourceGroup,
		Location:     "eastus",
	}
}

func (d DatabaseServerPair) ServerPairsConfig() interface{} {
	return map[string]interface{}{d.ServerPairTag: d}
}

func (d DatabaseServerPair) ReconfigureCSBWithServerDetails() {
	helpers.SetBrokerEnvAndRestart(
		helpers.EnvVar{Name: "MSSQL_DB_FOG_SERVER_PAIR_CREDS", Value: d.ServerPairsConfig()},
		helpers.EnvVar{Name: "GSB_SERVICE_CSB_AZURE_MSSQL_DB_FAILOVER_GROUP_PROVISION_DEFAULTS", Value: map[string]interface{}{"server_credential_pairs": d.ServerPairsConfig()}},
	)
}
