# Server configuration for {{.Questions.name}}
name: {{.Questions.name}}
environment: {{.Questions.environment}}
target: {{.Questions.target}}

# Server settings
server:
  host: 0.0.0.0
  port: {{if eq .Questions.environment "production"}}8443{{else}}8080{{end}}
  protocol: {{if eq .Questions.environment "production"}}https{{else}}http{{end}}
  ssl:
    enabled: {{if eq .Questions.environment "production"}}true{{else}}false{{end}}
    cert_file: "/etc/ssl/{{.Questions.name}}.crt"
    key_file: "/etc/ssl/{{.Questions.name}}.key"

# Performance settings
performance:
  max_connections: {{if eq .Questions.environment "production"}}1000{{else}}100{{end}}
  timeout: 30s
  keep_alive: true
  workers: {{if eq .Questions.environment "production"}}4{{else}}2{{end}}

# Security
security:
  cors:
    enabled: true
    origins: ["{{.Questions.environment}}.{{.Questions.name}}.com"]
  rate_limiting:
    enabled: {{if eq .Questions.environment "production"}}true{{else}}false{{end}}
    requests_per_minute: 60

# Monitoring
monitoring:
  health_check_path: "/health"
  metrics_path: "/metrics"
  log_level: {{if eq .Questions.environment "production"}}info{{else}}debug{{end}}