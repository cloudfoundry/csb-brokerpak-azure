# Consuming Services

## AI access family

The AI offerings in this brokerpak are:

- `azure_openai_key`
- `azure_foundry_identity`

`azure_openai_key` provisions an Azure OpenAI resource and returns the resource
API key plus endpoint during bind. It is the generic key-backed Azure offering
and is no longer split into model-pinned catalog plans.

`azure_foundry_identity` is currently a preview Foundry-aligned service family.
It provisions a backing Azure OpenAI resource with a single model deployment
and returns the OpenAI API key, endpoint, deployment metadata, and Foundry hub
and project metadata at bind time. This remains a transitional implementation:
the family is distinct in the catalog today, but the returned credentials are
still key-backed until true Foundry RBAC-backed credentials are wired in.

Documentation on all the services and plans can be found in the [Tanzu Cloud Service Broker for Azure documentation](https://docs.vmware.com/en/Tanzu-Cloud-Service-Broker-for-Azure/1.2/csb-azure/GUID-index.html).

Use this brokerpak README as the entry point for the current Azure AI access-family guidance.
