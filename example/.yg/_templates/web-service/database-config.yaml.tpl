# Database configuration for {{.Questions.name}}
name: {{.Questions.name}}
environment: {{.Questions.environment}}
target: {{.Questions.target}}

# Database connection
database:
  type: postgresql
  host: {{.Questions.name}}-db.{{.Questions.environment}}.local
  port: 5432
  name: {{.Questions.name}}_{{.Questions.environment}}
  username: {{.Questions.name}}_user
  ssl_mode: {{if eq .Questions.environment "production"}}require{{else}}disable{{end}}
  timezone: UTC

# Connection pool
pool:
  max_connections: {{if eq .Questions.environment "production"}}20{{else}}5{{end}}
  min_connections: {{if eq .Questions.environment "production"}}5{{else}}1{{end}}
  max_idle_time: 300s
  max_lifetime: 3600s

# Backup settings
backup:
  enabled: {{if eq .Questions.environment "production"}}true{{else}}false{{end}}
  schedule: "0 2 * * *"
  retention_days: {{if eq .Questions.environment "production"}}30{{else}}7{{end}}
  location: "s3://{{.Questions.name}}-backups/{{.Questions.environment}}"

# Monitoring
monitoring:
  slow_query_threshold: 1s
  log_queries: {{if eq .Questions.environment "development"}}true{{else}}false{{end}}
  metrics_enabled: true