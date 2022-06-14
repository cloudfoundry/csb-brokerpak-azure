package environment

import "csbbrokerpakazure/acceptance-tests/helpers/random"

// Server is for setting MSSQL_DB_SERVER_CREDS when creating CSB
func (m Metadata) Server() Server {
	return Server{
		Tag:           random.Name(random.WithMaxLength(10)),
		Name:          m.PreProvisionedSQLServer,
		ResourceGroup: m.ResourceGroup,
		AdminUsername: m.PreProvisionedSQLUsername,
		AdminPassword: m.PreProvisionedSQLPassword,
	}
}

type Server struct {
	Tag           string `json:""`
	Name          string `json:"server_name"`
	ResourceGroup string `json:"server_resource_group"`
	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`
}
