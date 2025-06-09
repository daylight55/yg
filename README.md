# yg

A CLI tool to generate files from templates based on interactive prompts. Create configuration files, API specifications, and more using customizable templates.

## Features

- **Interactive Prompts**: Choose options using arrow keys, space for multiple selection
- **Searchable Interface**: Filter choices with peco-like search functionality  
- **Dynamic Questions**: Questions adapt based on previous answers
- **Template Engine**: Flexible template rendering with Go templates (YAML, JSON, XML, etc.)
- **Multi-output**: Generate files for multiple environments and targets
- **CLI Options**: Skip prompts with command-line flags
- **Signal Handling**: Graceful shutdown with Ctrl+C
- **Directory Templates**: Support for multi-file template directories 🆕
- **File Templates**: Traditional single-file template support (backward compatible)
- **Template Question Control**: Configure which question determines template selection 🆕
- **CLI Example Output**: Shows equivalent CLI command after interactive session 🆕

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

### Demo

See yg in action:

![yg demo](https://github.com/user-attachments/assets/demo-placeholder.gif)

*インタラクティブなプロンプトとリアルタイム検索機能のデモンストレーション*

### Interactive Mode

```bash
yg
```

This will prompt you with questions to:

1. Select template type (web-service, configuration, batch-job)
2. Choose item name (with search)
3. Select environments (multiple choice)
4. Choose target destinations (based on selected environments)
5. Preview generated files
6. Confirm generation
7. Show CLI command example for next time 🆕

After completion, yg displays the equivalent CLI command:

```bash
CLI Example:
yg --yes --answer templateType=web-service --answer name=user-service --answer environment=development --answer target=dev-region-1
```

### CLI Mode

```bash
# Generate configuration file
yg --answer templateType=configuration --answer name=my-config --answer environment=development,staging --answer target=dev-region-1,staging-region-1 --yes

# Generate batch job  
yg --answer templateType=batch-job --answer name=data-processor --answer environment=production --answer target=prod-region-1 --yes

# Generate web service (directory template)
yg --answer templateType=web-service --answer name=user-service --answer environment=development --answer target=dev-region-1 --yes
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
    ├── config.yaml         # Question and template configuration
    ├── configuration.yaml  # Single file template
    ├── batch-job.yaml     # Single file template
    └── web-service/       # Directory template (new feature)
        ├── .template-config.yaml
        ├── server-config.yaml
        ├── database-config.yaml
        ├── api-spec.yaml
        └── monitoring-config.yaml
```

### Configuration File (`.yg-config.yaml`)

The configuration file supports both file and directory templates:

```yaml
# Template definitions (new feature)
templates:
  web-service:
    type: directory      # Directory template with multiple files
    path: web-service
  configuration:
    type: file          # Single file template (traditional)
    path: configuration.yaml
  batch-job:
    type: file
    path: batch-job.yaml

# Question configuration
questions:
  template_question: "templateType"  # Which question determines template selection 🆕
  order:                    # Question execution order 🆕
    - templateType
    - name
    - environment
    - target
  definitions:             # Question definitions 🆕
    templateType:
      prompt: "What type of template do you want to use?"
      choices:
        - web-service   # Directory template
        - configuration    # File template
        - batch-job          # File template
    name:
      prompt: "What is the name of your item?"
      type:
        dynamic:
          dependency_questions: ["templateType"]
        interactive: true
      choices:
        web-service:
          - user-service
          - payment-api
        configuration:
          - database-config
          - cache-config
        batch-job:
          - data-processor
          - report-generator
    environment:
      prompt: "Which environment do you want to target?"
      type:
        multiple: true
      choices:
        - development
        - staging
        - production
    target:
      prompt: "Which target destinations do you want to deploy to?"
      type:
        multiple: true
        dynamic:
          dependency_questions: ["environment"]
      choices:
        development:
          - dev-region-1
          - dev-region-2
        staging:
          - staging-region-1
        production:
          - prod-region-1
```

### Template Files

#### Single File Templates (Traditional)

Templates use Go template syntax and have metadata headers:

```yaml
path: {{.Questions.environment}}/{{.Questions.target}}/configs
filename: {{.Questions.name}}-config.yaml
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

**Files** 

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

## Creating Demo GIF

To record and update the demo GIF shown above, follow these steps:

### Prerequisites

Install a terminal recording tool:

```bash
# Using asciinema + agg for high-quality GIF
npm install -g @asciinema/cli
npm install -g @asciinema/agg

# Or using terminalizer
npm install -g terminalizer

# Or using ttygif (requires ttyrec)
brew install ttyrec
git clone https://github.com/icholy/ttygif.git
cd ttygif && make && sudo make install
```

### Recording Steps

1. **Prepare the environment:**
   ```bash
   # Build the tool
   go build -o yg
   
   # Set up a clean example directory
   cp -r example/ demo-example/
   cd demo-example/
   ```

2. **Record the session:**
   ```bash
   # Using asciinema (recommended)
   asciinema rec yg-demo.cast
   
   # Or using terminalizer
   terminalizer record yg-demo
   
   # Or using ttygif
   ttyrec yg-demo.tty
   ```

3. **Run the demo script:**
   ```bash
   # Start recording and run:
   ../yg
   
   # In the interactive session, demonstrate:
   # 1. Select template type (e.g., microservice)
   # 2. Use search functionality for app name (e.g., type "api" and select "sample-api-1")
   # 3. Select environment (e.g., dev)
   # 4. Select multiple clusters (e.g., dev-cluster-1, dev-cluster-2)
   # 5. Review the output preview
   # 6. Confirm generation
   # 7. Show the CLI example output
   
   # Stop recording (Ctrl+C or exit)
   ```

4. **Convert to GIF:**
   ```bash
   # Using asciinema + agg
   agg yg-demo.cast yg-demo.gif
   
   # Using terminalizer
   terminalizer render yg-demo
   
   # Using ttygif
   ttygif yg-demo.tty
   ```

5. **Optimize the GIF:**
   ```bash
   # Reduce file size if needed
   gifsicle -O3 --resize-width 800 yg-demo.gif -o yg-demo-optimized.gif
   ```

### Upload to GitHub

1. **Create a new issue or comment** in this repository
2. **Drag and drop the GIF** into the comment box
3. **Copy the generated URL** (format: `https://github.com/user-attachments/assets/[hash].gif`)
4. **Update this README** by replacing `demo-placeholder.gif` with the actual URL:
   ```markdown
   ![yg demo](https://github.com/user-attachments/assets/[actual-hash].gif)
   ```

### Demo Script Guidelines

The ideal demo should showcase:

- **Interactive prompts** with arrow key navigation
- **Real-time search functionality** (typing to filter choices)
- **Multiple selection** capabilities (spacebar to select multiple items)
- **Dynamic questions** (how choices change based on previous answers)
- **Output preview** showing generated file paths
- **CLI example output** at the end
- **Clean, readable terminal output** with appropriate timing

Keep the recording under 30 seconds for better loading performance.

## License

MIT License - see [LICENSE](LICENSE) file for details.
