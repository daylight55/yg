# API Specification for {{.Questions.name}}
openapi: 3.0.3
info:
  title: {{.Questions.name}} API
  description: REST API for {{.Questions.name}} service
  version: 1.0.0
  contact:
    name: {{.Questions.name}} Team
    email: team@{{.Questions.name}}.com

servers:
  - url: {{if eq .Questions.environment "production"}}https{{else}}http{{end}}://{{.Questions.name}}.{{.Questions.environment}}.{{.Questions.target}}.com
    description: {{.Questions.environment}} environment

paths:
  /health:
    get:
      summary: Health check endpoint
      responses:
        '200':
          description: Service is healthy
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                    example: healthy
                  timestamp:
                    type: string
                    format: date-time

  /{{.Questions.name}}:
    get:
      summary: Get {{.Questions.name}} data
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            default: 10
            maximum: 100
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      type: object
                  total:
                    type: integer
                  environment:
                    type: string
                    example: {{.Questions.environment}}

components:
  securitySchemes:
    {{if eq .Questions.environment "production"}}BearerAuth:
      type: http
      scheme: bearer{{else}}ApiKey:
      type: apiKey
      in: header
      name: X-API-Key{{end}}

security:
  {{if eq .Questions.environment "production"}}- BearerAuth: []{{else}}- ApiKey: []{{end}}