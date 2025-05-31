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
yg --app deployment --name my-app --env dev,staging --cluster dev-cluster-1,staging-cluster-1 --yes

# Generate job YAML  
yg --app job --name batch-job --env production --cluster prod-cluster-1 --yes
```

### CLI Options

- `--app`: Application type (deployment, job)
- `--name`: Application name
- `--env`: Environments (comma-separated for multiple)
- `--cluster`: Clusters (comma-separated for multiple)
- `--yes`: Skip confirmation prompts

## Configuration

### Directory Structure

```console
.yg/
└── _templates/
    ├── .yg-config.yaml    # Question configuration
    ├── deployment.yaml    # Deployment template
    └── job.yaml          # Job template
```

### Configuration File (`.yg-config.yaml`)

```yaml
questions:
  app:
    prompt: "アプリの種類はなんですか？"
    choices:
      - deployment
      - job
  appName:
    prompt: "アプリ名は何ですか？"
    type:
      dynamic:
        dependency_questions: ["app"]
      interactive: true
    choices:
      deployment:
        - sample-server-1
        - sample-server-2
      job:
        - sample-job-1
        - sample-job-2
  env:
    prompt: "環境名はなんですか？"
    multiple: true
    choices:
      - dev
      - staging
      - production
  cluster:
    prompt: "クラスターはどこですか？"
    multiple: true
    type:
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

## Examples

### Example Output

Running `yg --app deployment --name my-app --env dev --cluster dev-cluster-1 --yes` generates:

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
