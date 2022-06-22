// Portions Copyright 2020 Pivotal Software, Inc.
// Portions Copyright 2020 Service Broker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http:#www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/denisenkom/go-mssqldb"
)

var db *sql.DB

var server string
var port int
var user string
var password string
var database string
var query string
var err error

func main() {
	if len(os.Args) < 5 {
		log.Fatal("Usage: psqlcmd <hostname> <port> <username> [password] <database> <query>")
	}

	server = os.Args[1]
	port, err = strconv.Atoi(os.Args[2])
	user = os.Args[3]

	if len(os.Args) > 6 {
		password = os.Args[4]
		database = os.Args[5]
		query = os.Args[6]
	} else {
		ok := false
		password, ok = os.LookupEnv("MSSQL_PASSWORD")
		if !ok {
			log.Fatal("Usage: psqlcmd <hostname> <port> <username> [password] <database> <query> - no password provided on command line or in environment variable MSSQL_PASSWORD")
		}
		database = os.Args[4]
		query = os.Args[5]
	}

	// Build connection string
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
		server, user, password, port, database)

	var err error

	// Create connection pool
	db, err = sql.Open("sqlserver", connString)
	if err != nil {
		log.Fatal("Error creating connection pool: ", err.Error())
	}
	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	tsql := fmt.Sprintf(query)

	_, err = db.ExecContext(ctx, tsql)
	if err != nil {
		log.Fatal(err)
	}
}
