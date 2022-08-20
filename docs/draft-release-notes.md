## Release notes for next release:

### New feature:
- Client Secret field validation in failover service: We introduce a less restrictive validation. The main validation is on the supplier's side, and we just check if the field is empty or too long.
### Fix:
- The value of the Client Secret is considered a sensible field. We avoid logging its value in plain text in our custom terraform provider csbmssqldbrunfailover.
