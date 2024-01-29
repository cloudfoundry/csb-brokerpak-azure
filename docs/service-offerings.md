
# Service offering and plans docs

Documentation on all the services and plans can be found [here](https://docs.vmware.com/en/Tanzu-Cloud-Service-Broker-for-Azure/1.6/csb-azure/GUID-index.html).

# CSB Configuration 

Some services have extra configuration when setting up the broker.

## Azure SQL databases on pre-configured database servers *csb-azure-mssql-db*

*csb-azure-mssql-db* manages Azure SQL databases on pre-configured database servers on Azure.

### Configuring Global Defaults

An operator will likely configure *server_credentials* for developers to use.

See [configuration documentation](./configuration.md) and [Azure installation documentation](azure-installation.md) for reference.

To globally configure *server_credential*, include the following in the configuration file for the broker:

```yaml
azure:
  mssql_db_server_creds: '{ 
        "server1": { 
            "admin_username":"...", 
            "admin_password":"...", 
            "server_name":"...", 
            "server_resource_group":..."
          },
          "server2": {
            "admin_username":"...",
            ...
          }
      }' 
```

A developer could create a new failover group database on *server1* like this:
```bash
cf create-service csb-azure-mssql-db medium medium-sql -c '{"server":"server1"}'
```
