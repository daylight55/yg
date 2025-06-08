# yg

YAML template generator - A CLI tool to generate YAML files from templates based on interactive prompts.

## Features

- **Interactive Prompts**: Choose options using arrow keys, space for multiple selection
- **Searchable Interface**: Filter choices with peco-like search functionality  
- **Dynamic Questions**: Questions adapt based on previous answers
- **Template Engine**: Flexible YAML template rendering with Go templates
- **Multi-output**: Generate files for multiple environments and clusters
- **CLI Options**: Skip prompts with command-line flags
- **Signal Handling**: Graceful shutdown with Ctrl+C
- **Directory Templates**: Support for multi-file template directories 🆕
- **File Templates**: Traditional single-file template support (backward compatible)
- **Template Question Control**: Configure which question determines template selection 🆕

## Installation

### Using go install

```bash
go install github.com/daylight55/yg@latest
```

### Using Homebrew (when available)

```bash
brew install daylight55/yg/yg
```

### From Source

```bash
git clone https://github.com/daylight55/yg.git
cd yg
go build -o yg
```

## Usage

### Interactive Mode

```bash
yg
```

This will prompt you with questions to:

1. Select application type (deployment, job)
2. Choose application name (with search)
3. Select environments (multiple choice)
4. Choose clusters (based on selected environments)
5. Preview generated files
6. Confirm generation

### CLI Mode

```bash
# Generate deployment YAML
yg --answer app=deployment --answer appName=my-app --answer env=dev,staging --answer cluster=dev-cluster-1,staging-cluster-1 --yes

# Generate job YAML  
yg --answer app=job --answer appName=batch-job --answer env=production --answer cluster=prod-cluster-1 --yes

# Generate microservice YAML (directory template)
yg --answer app=microservice --answer appName=my-api --answer env=dev --answer cluster=dev-cluster-1 --yes
```

### CLI Options

- `--answer key=value`: Provide answers for questions (use multiple times for different questions)
- `--config`, `-c`: Path to config file (default: ./.yg/config.yaml or ./.yg/config.yml)
- `--yes`: Skip confirmation prompts

## Configuration

### Directory Structure

```console
.yg/
└── _templates/
    ├── .yg-config.yaml      # Question and template configuration
    ├── deployment.yaml      # Single file template
    ├── job.yaml            # Single file template
    └── microservice/       # Directory template (new feature)
        ├── .template-config.yaml
        ├── deployment.yaml
        ├── service.yaml
        ├── configmap.yaml
        └── ingress.yaml
```

### Configuration File (`.yg-config.yaml`)

The configuration file supports both file and directory templates:

```yaml
# Template definitions (new feature)
templates:
  microservice:
    type: directory      # Directory template with multiple files
    path: microservice
  deployment:
    type: file          # Single file template (traditional)
    path: deployment.yaml
  job:
    type: file
    path: job.yaml

# Question configuration
questions:
  template_question: "app"  # Which question determines template selection 🆕
  order:                    # Question execution order 🆕
    - app
    - appName
    - env
    - cluster
  definitions:             # Question definitions 🆕
    app:
      prompt: "アプリの種類はなんですか？"
      choices:
        - microservice   # Directory template
        - deployment    # File template
        - job          # File template
    appName:
      prompt: "アプリ名は何ですか？"
      type:
        dynamic:
          dependency_questions: ["app"]
        interactive: true
      choices:
        microservice:
          - sample-api-1
          - sample-api-2
        deployment:
          - sample-server-1
          - sample-server-2
        job:
          - sample-job-1
          - sample-job-2
    env:
      prompt: "環境名はなんですか？"
      type:
        multiple: true
      choices:
        - dev
        - staging
        - production
    cluster:
      prompt: "クラスターはどこですか？"
      type:
        multiple: true
        dynamic:
          dependency_questions: ["env"]
      choices:
        dev:
          - dev-cluster-1
          - dev-cluster-2
        staging:
          - staging-cluster-1
        production:
          - production-cluster-1
```

### Template Files

#### Single File Templates (Traditional)

Templates use Go template syntax and have metadata headers:

```yaml
path: {{.Questions.env}}/{{.Questions.cluster}}/deployment
filename: {{.Questions.appName}}-deployment.yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Questions.appName}}
  namespace: {{.Questions.env}}
spec:
  # ... template content
```

#### Directory Templates (New Feature)

Directory templates consist of multiple files with shared configuration:

**`.template-config.yaml`**:
```yaml
output:
  base_path: "{{.Questions.env}}/{{.Questions.cluster}}/{{.Questions.appName}}"

files:
  deployment.yaml:
    filename: "{{.Questions.appName}}-deployment.yaml"
  service.yaml:
    filename: "{{.Questions.appName}}-service.yaml"
  configmap.yaml:
    filename: "{{.Questions.appName}}-configmap.yaml"
  ingress.yaml:
    filename: "{{.Questions.appName}}-ingress.yaml"
```

Each file in the directory is a regular Go template without metadata headers.

### Template Question Configuration 🆕

The `template_question` field allows you to explicitly specify which question determines the template selection:

```yaml
questions:
  template_question: "serviceType"  # Specify which question provides template name
  order:
    - region                        # First question (multiple selection)
    - serviceType                   # This determines template
    - environment                   # Last question (multiple selection)
  definitions:
    region:
      prompt: "Select regions"
      type:
        multiple: true
      choices: ["us-east", "us-west", "eu-west"]
    serviceType:
      prompt: "Service type?"
      choices: ["web-app", "api-service", "batch-job"]
    environment:
      prompt: "Target environments"
      type:
        multiple: true
      choices: ["dev", "staging", "prod"]
```

**Benefits:**
- **Flexible ordering**: Template-determining question doesn't need to be first
- **Clear configuration**: Explicit rather than heuristic-based template selection
- **Backward compatible**: Falls back to original heuristic if not specified

**Fallback behavior**: If `template_question` is not specified, the system uses the first non-multiple question in order (original behavior).

## Examples

### Example Outputs

#### Single File Template Output

Running `yg --answer app=deployment --answer appName=my-app --answer env=dev --answer cluster=dev-cluster-1 --yes` generates:

**File**: `dev/dev-cluster-1/deployment/my-app-deployment.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: dev
  labels:
    app: my-app
    env: dev
    cluster: dev-cluster-1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
        env: dev
        cluster: dev-cluster-1
    spec:
      containers:
      - name: my-app
        image: nginx:latest
        ports:
        - containerPort: 80
        env:
        - name: ENV
          value: dev
        - name: CLUSTER
          value: dev-cluster-1
```

#### Directory Template Output

Running `yg --answer app=microservice --answer appName=my-api --answer env=dev --answer cluster=dev-cluster-1 --yes` generates multiple files:

**Files**: 
- `dev/dev-cluster-1/my-api/my-api-deployment.yaml`
- `dev/dev-cluster-1/my-api/my-api-service.yaml`
- `dev/dev-cluster-1/my-api/my-api-configmap.yaml`
- `dev/dev-cluster-1/my-api/my-api-ingress.yaml`

This allows you to generate complete microservice manifests in one command.

## Development

### Prerequisites

- Go 1.21 or later

### Setup

```bash
git clone https://github.com/daylight55/yg.git
cd yg
go mod tidy
```

### Build and Test

```bash
# Build
go build -o yg

# Run tests
go test ./...

# Check coverage
go test -cover ./...

# Lint
golangci-lint run
```

### Project Structure

```console
├── cmd/                    # CLI commands
├── internal/
│   ├── config/            # Configuration management
│   ├── generator/         # Main generation logic
│   ├── prompt/           # Interactive prompts
│   └── template/         # Template processing
├── .github/workflows/     # CI/CD
├── .yg/_templates/       # Sample configuration and templates
└── main.go              # Entry point
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure tests pass and coverage remains high
6. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.
