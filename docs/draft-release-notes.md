## Release notes for next release:

### New feature:

### Fix:
- **Failover service client secret parameter is no longer exposed:** The value of the Client Secret is considered a sensible field. We avoid logging its value in plain text in our custom terraform provider csbmssqldbrunfailover.
- **Failover service parameters validation is more flexible:** Previously, the validation of the parameters was carried out in a more restrictive way adding rigidity to the admitted values. With this change, the main validation is on the supplier side, and we only check if the field is empty.
