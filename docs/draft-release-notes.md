## Release notes for next release:

### New feature:
- Use text field in the location property
- default Terraform version now 1.1.6 and upgrade path added
- new MS-SQL Server Terraform provider used for bindings

### Fix:
- minimum constraint on PostreSQL storage_gb is now enforced
- adds lifecycle.prevent_destroy to all data services to provide extra layer of protection against data loss
