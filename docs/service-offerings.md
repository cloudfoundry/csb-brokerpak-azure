# Service offering and plans docs

## AI access family

The AI services currently exposed in this brokerpak are `azure_openai_key` and
`azure_foundry_identity`.

Operator guidance:

- Choose `azure_openai_key` when a dedicated Azure OpenAI resource key is the intended product.
- Choose `azure_foundry_identity` when you need the distinct catalog family for Foundry hub, project, and per-model access tracking.
- Treat `azure_foundry_identity` as a transitional preview: it currently provisions a backing Azure OpenAI deployment and returns key-backed credentials plus Foundry metadata.
- Keep key-backed and identity-backed AI services as separate catalog families.

Documentation on all the services and plans can be found in the [Tanzu Cloud Service Broker for Azure service docs](https://docs.vmware.com/en/Tanzu-Cloud-Service-Broker-for-Azure/1.6/csb-azure/GUID-index.html).

## CSB Configuration

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
