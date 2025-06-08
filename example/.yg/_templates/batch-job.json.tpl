path: {{.Questions.environment}}/{{.Questions.target}}/jobs
filename: {{.Questions.name}}-job.json
---
{
  "name": "{{.Questions.name}}",
  "environment": "{{.Questions.environment}}",
  "target": "{{.Questions.target}}",
  "type": "batch-job",
  "schedule": "0 2 * * *",
  "timeout": "3600s",
  "retries": 3,
  "parameters": {
    "input_path": "/data/{{.Questions.environment}}/{{.Questions.name}}/input",
    "output_path": "/data/{{.Questions.environment}}/{{.Questions.name}}/output",
    "batch_size": {{if eq .Questions.environment "production"}}1000{{else}}100{{end}},
    "parallel_workers": {{if eq .Questions.environment "production"}}4{{else}}2{{end}}
  },
  "resources": {
    "cpu": "{{if eq .Questions.environment \"production\"}}2{{else}}1{{end}}",
    "memory": "{{if eq .Questions.environment \"production\"}}4Gi{{else}}2Gi{{end}}",
    "disk": "10Gi"
  },
  "notifications": {
    "on_success": {{if eq .Questions.environment "production"}}true{{else}}false{{end}},
    "on_failure": true,
    "channels": ["email", "slack"]
  }
}
