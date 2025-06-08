# Monitoring configuration for {{.Questions.name}}
name: {{.Questions.name}}
environment: {{.Questions.environment}}
target: {{.Questions.target}}

# Metrics collection
metrics:
  enabled: true
  port: 9090
  path: "/metrics"
  interval: {{if eq .Questions.environment "production"}}15s{{else}}30s{{end}}
  retention: {{if eq .Questions.environment "production"}}30d{{else}}7d{{end}}

# Health checks
health_checks:
  liveness:
    path: "/health/live"
    interval: 10s
    timeout: 5s
    failure_threshold: 3
  readiness:
    path: "/health/ready"
    interval: 5s
    timeout: 3s
    failure_threshold: 1

# Logging
logging:
  level: {{if eq .Questions.environment "production"}}info{{else}}debug{{end}}
  format: json
  output: stdout
  structured: true
  fields:
    service: {{.Questions.name}}
    environment: {{.Questions.environment}}
    target: {{.Questions.target}}

# Alerting
alerts:
  enabled: {{if eq .Questions.environment "production"}}true{{else}}false{{end}}
  channels:
    - email
    - slack
  thresholds:
    error_rate: 0.05
    response_time_p95: 2s
    cpu_usage: 80
    memory_usage: 85

# Tracing
tracing:
  enabled: {{if eq .Questions.environment "production"}}true{{else}}false{{end}}
  sample_rate: {{if eq .Questions.environment "production"}}0.1{{else}}1.0{{end}}
  endpoint: "http://jaeger.{{.Questions.environment}}.local:14268/api/traces"