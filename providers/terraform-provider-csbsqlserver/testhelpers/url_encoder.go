package testhelpers

import (
	"fmt"
	"maps"
	"net/url"
)

const (
	queryParamDatabaseKey = "database"
	queryParamEncryptKey  = "encrypt"
)

type Encoder struct {
	server      string
	username    string
	password    string
	port        int
	queryParams map[string]string
}

func NewEncoder(
	server,
	username,
	password,
	database,
	encrypt string,
	port int,
) *Encoder {
	queryParams := map[string]string{queryParamDatabaseKey: database, queryParamEncryptKey: encrypt}
	return &Encoder{
		server:      server,
		username:    username,
		password:    password,
		port:        port,
		queryParams: queryParams,
	}
}

func (b *Encoder) Encode() string {
	u := createURL(b.server, b.username, b.password, b.port)
	u.RawQuery = createQueryParams(b.queryParams).Encode()

	return u.String()
}

func (b *Encoder) EncodeWithoutDB() string {
	qp := make(map[string]string)
	maps.Copy(qp, b.queryParams)
	delete(qp, queryParamDatabaseKey)
	u := createURL(b.server, b.username, b.password, b.port)
	u.RawQuery = createQueryParams(qp).Encode()

	return u.String()
}

func createQueryParams(params map[string]string) url.Values {
	q := url.Values{}
	for key, value := range params {
		q.Add(key, value)
	}
	return q
}

func createURL(server, username, password string, port int) url.URL {
	const scheme = "sqlserver"
	u := url.URL{
		Scheme: scheme,
		User:   url.UserPassword(username, password),
		Host:   fmt.Sprintf("%s:%d", server, port),
	}

	return u
}
