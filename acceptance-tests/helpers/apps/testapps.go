package apps

import (
	"fmt"
	"os"
)

type AppCode string

const (
	Cosmos     AppCode = "cosmosdbapp"
	Storage    AppCode = "storageapp"
	MongoDB    AppCode = "mongodbapp"
	MySQL      AppCode = "mysqlapp"
	MSSQL      AppCode = "mssqlapp"
	PostgreSQL AppCode = "postgresqlapp"
	Redis      AppCode = "redisapp"
)

func (a AppCode) Dir() string {
	for _, d := range []string{"apps", "../apps"} {
		p := fmt.Sprintf("%s/%s", d, string(a))
		_, err := os.Stat(p)
		if err == nil {
			return p
		}
	}

	panic(fmt.Sprintf("could not find source for app: %s", a))
}

func WithApp(app AppCode) Option {
	switch app {
	case Cosmos, Storage:
		return WithOptions(WithDir(app.Dir()), WithMemory("100MB"), WithDisk("250MB"))
	default:
		return WithOptions(WithPreBuild(app.Dir()), WithMemory("100MB"), WithDisk("250MB"))
	}
}
