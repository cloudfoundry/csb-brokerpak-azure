locals {
  ttl_expires_at = timeadd(timestamp(), "${var.ttl_hours}h")
  api_version    = "2024-02-01"

  # Truncate instance_name to satisfy Azure 24-char Cognitive Services limit.
  account_name = substr(replace(var.instance_name, "-", ""), 0, 24)

  common_tags = merge(var.labels, {
    TTLExpiry   = local.ttl_expires_at
    ManagedBy   = "cloud-service-broker"
    Environment = "sandbox"
  })

  deployments = jsonencode([
    {
      name     = "gpt-35-turbo"
      model    = "gpt-35-turbo"
      version  = "0125"
      capacity = var.gpt35_capacity
    },
    {
      name     = "text-embedding-ada-002"
      model    = "text-embedding-ada-002"
      version  = "2"
      capacity = var.embedding_capacity
    },
  ])
}
