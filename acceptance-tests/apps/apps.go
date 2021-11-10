package apps

import "fmt"

type AppCode string

const (
	Cosmos    AppCode = "cosmosdbapp"
	Storage   AppCode = "storageapp"
	MongoDB   AppCode = "mongodbapp"
	MySQL     AppCode = "mysqlapp"
	MSSQL      AppCode = "mssqlapp"
	PostgreSQL AppCode = "postgresqlapp"
	Redis      AppCode = "redisapp"
)

func (a AppCode) Dir() string {
	return fmt.Sprintf("../apps/%s", string(a))
}
