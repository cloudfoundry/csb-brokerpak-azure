# Creating an Azure MySQL DB for Service Broker State

The Cloud Service Broker (CSB) requires a MySQL database to keep its internal state.

We recommend using a native MySQL instance for the CSB state database. 
Follow the [Azure documentation](https://azure.microsoft.com/en-us/products/mysql/) to create a MySQL database instance in Azure.

The CSB requires the following credentials to connect to the MySQL database:

```bash
Server Details
FQDN: <name>.mysql.database.azure.com
Admin Username: ifOuuVydAjJNzYHF@<name>
Admin Password: 9Kagdsl8VWhw1eQpp8WMVQplnOp156Ly
Database Name: csb-db
```

If you're `cf push`ing the broker, these values should be used for the config file values:
* db.host
* db.user
* db.password
* db.name

or the environment variables:
* DB_HOST
* DB_USERNAME
* DB_PASSWORD
* DB_NAME

If you're deploying the broker as a tile through OpsMan, these values should be used for the following fields on the *Cloud Service Broker for Microsoft Azure -> Service Broker Config* tab:
* Database host
* Database username
* Database password
* Database name


