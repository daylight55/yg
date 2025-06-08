path: {{.Questions.environment}}/{{.Questions.target}}/configs
filename: {{.Questions.name}}-config.yaml
---
# Configuration file for {{.Questions.name}}
name: {{.Questions.name}}
environment: {{.Questions.environment}}
target: {{.Questions.target}}

# Application settings
app:
  name: {{.Questions.name}}
  version: "1.0.0"
  debug: {{if eq .Questions.environment "development"}}true{{else}}false{{end}}

# Database configuration
database:
  host: {{.Questions.name}}-db.{{.Questions.environment}}.local
  port: 5432
  name: {{.Questions.name}}_{{.Questions.environment}}
  ssl: {{if eq .Questions.environment "production"}}require{{else}}disable{{end}}

# Service configuration
service:
  port: 8080
  timeout: 30s
  max_connections: 100

# Logging
logging:
  level: {{if eq .Questions.environment "production"}}info{{else}}debug{{end}}
  format: json
  output: stdout
