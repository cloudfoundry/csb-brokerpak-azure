# Billing

The Service Broker automatically labels supported resources with organization GUID, space GUID and instance ID.

When these supported services are provisioned, they will have the following labels populated with information from the request:

 * `pcf-organization-guid`
 * `pcf-space-guid`
 * `pcf-instance-id`

Labels have a more restricted character set than the Service Broker so unsupported characters will be mapped to the underscore character (`_`).

## Support

All brokerpaks should support these billing tags.

