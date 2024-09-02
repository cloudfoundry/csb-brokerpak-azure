package apps

import (
	"csbbrokerpakazure/acceptance-tests/helpers/testpath"
)

type AppCode string

const (
	MongoDB    AppCode = "mongodbapp"
	MSSQL      AppCode = "mssqlapp"
	PostgreSQL AppCode = "postgresqlapp"
	Redis      AppCode = "redisapp"
)

func (a AppCode) Dir() string {
	return testpath.BrokerpakFile("acceptance-tests", "apps", string(a))
}

func WithApp(app AppCode) Option {
	return WithOptions(WithPreBuild(app.Dir()), WithMemory("100MB"), WithDisk("250MB"))
}
