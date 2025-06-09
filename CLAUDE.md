# Template Generator

This project is a Go language tool that generates template files based on interactive prompts. It can generate configuration files, API specifications, and other template-based content.

## Recent Updates

- Added preview control functionality (Issue #22)
  - Config file setting: `preview.enabled`
  - CLI option: `--no-preview`
  - CLI option takes precedence over config setting
  - Default behavior: preview enabled

## Claude Rule

- Respond in Japanese
- Update README.md/CLAUDE.md after modifications
- Run tests and lint after modifications to ensure success, fix if they fail
- Finally execute Git commit and push to GitHub

## Guidelines

- Style: Develop using Go language standard project structure.
- Coding: Focus on writing readable code.
  - Reference: https://google.github.io/styleguide/go/guide
- Documentation: Create comments that can be output with godoc.
- Test: Create test code with 90% or higher test coverage. Summarize test cases in Markdown in the test directory.
- CI/CD: Execute tests with CI and go releaser with CD using GitHub Actions, and distribute binaries on GitHub. Make it installable with go install.
- Quality: Handle errors appropriately. Always create test cases for abnormal systems, and ensure inappropriate output is not executed when problems occur.
- CommitMessage: Use Semantic Commit for commit messages, and increase semantic version when merging to main branch.

## Tool Design

The purpose is to output template files according to environment directories and target directories, based on templates that match the type of content.
Users can simply answer prompts to output files that match templates to specified directories.

ポイントは以下である。

- Be able to specify output directories according to prompts.
- Be able to select template files to use according to prompts.
- Prompts can be dynamically selected according to previous answers.
- Prompt answers can accept string input and allow interactive search with real-time search.
- Interactive search function filters choices in real-time when characters are entered, supporting both exact match and partial match.
- Be able to embed values received from prompts into template files according to specified template paths.
- Multiple selection is possible for output directory question prompts.
- Prompt question answers can also be set with CLI options (--answer key=value format).
- Can be canceled with Ctrl + C during prompts. Return to original state when receiving forced termination signal.
- Provide an attractive Terminal UI.

### Usage

### Ideal directory structure

```console
├── .yg
│   └── _templates
│       ├── .yg_config.yaml
│       ├── deployment.yaml
│       └── job.yaml
├── dev
│   ├── dev-cluster-1
│   ├── dev-cluster-2
│   └── dev-cluster-3
├── production
│   ├── production-cluster-1
│   └── production-cluster-2
└── staging
    ├── staging-cluster-1
    ├── staging-cluster-2
    └── staging-cluster-3
```

### Command

```console
# コマンド実行
$ yaml-generator

# プロンプトが出る。矢印キーで選択し、スペースキーでチェック、Enterで次に進める。
Q. アプリの種類はなんですか？
[x] microservice   # ディレクトリテンプレート（複数ファイル一括生成）
[] deployment      # 単一ファイルテンプレート
[] job             # 単一ファイルテンプレート

# Interactive search機能：リアルタイム検索で選択肢をフィルタリング
Q. アプリ名は何ですか？ (入力で検索、↓↑で選択):
> sample-api-1
sample-api-1  # 完全一致している選択肢がハイライト表示

# または部分文字列でリアルタイムフィルタリング
Q. アプリ名は何ですか？ (入力で検索、↓↑で選択):
> api
sample-api-1
sample-api-2  # "api"を含む選択肢のみリアルタイム表示
sample-api-3

# 文字を入力するとリアルタイムで選択肢が絞り込まれ、矢印キーで選択してEnterで決定

# プロンプトが出る。矢印キーで選択し、スペースキーでチェック、Enterで次に進める。
Q. 環境名はなんですか？
[x] dev
[] staging
[] production

# 上記で選択した環境に合わせたクラスターの選択肢を出力し、複数選択に対応する。
Q. クラスターはどこですか？
[x] dev-cluster-1
[x] dev-cluster-2
[x] dev-cluster-3

# 出力先ディレクトリと、出力するYAMLのレンダリング想定を出力し、確認を促す。

Output:

# ディレクトリテンプレート（microservice）選択時の複数ファイル出力例
* dev/dev-cluster-1/sample-api-1/sample-api-1-deployment.yaml
* dev/dev-cluster-1/sample-api-1/sample-api-1-service.yaml
* dev/dev-cluster-1/sample-api-1/sample-api-1-configmap.yaml
* dev/dev-cluster-1/sample-api-1/sample-api-1-ingress.yaml

dev/dev-cluster-2/sample-api-1/sample-api-1-deployment.yaml
dev/dev-cluster-2/sample-api-1/sample-api-1-service.yaml
dev/dev-cluster-2/sample-api-1/sample-api-1-configmap.yaml
dev/dev-cluster-2/sample-api-1/sample-api-1-ingress.yaml

dev/dev-cluster-3/sample-api-1/sample-api-1-deployment.yaml
dev/dev-cluster-3/sample-api-1/sample-api-1-service.yaml
dev/dev-cluster-3/sample-api-1/sample-api-1-configmap.yaml
dev/dev-cluster-3/sample-api-1/sample-api-1-ingress.yaml


Q. 出力して問題ないですか? [Y/N]
> Y

# 生成成功の結果とCLI例が出力される。（新機能）
CLI Example:
yg --yes --answer app=microservice --answer appName=sample-api-1 --answer env=dev --answer cluster=dev-cluster-1,dev-cluster-2,dev-cluster-3

generated!
```

**拡張された`.yg-config.yaml`（ディレクトリテンプレート対応）**

```yaml
# テンプレート設定（新機能）
templates:
  microservice:
    type: directory   # ディレクトリテンプレート
    path: microservice
  deployment:
    type: file       # 単一ファイルテンプレート（従来）
    path: deployment.yaml
  job:
    type: file
    path: job.yaml

questions:
  # テンプレート名を決定する質問を明示的に指定（新機能）
  template_question: "app"
  # 質問の実行順序を明示的に指定
  order:
    - app
    - appName
    - env
    - cluster
  # 質問の定義
  definitions:
    app:
      prompt: "アプリの種類はなんですか？"
      choices:
        - microservice  # ディレクトリテンプレート
        - deployment    # 従来の単一ファイル
        - job          # 従来の単一ファイル
    appName:
      prompt: "アプリ名は何ですか？"
      type:
        dynamic:
          dependency_questions: ["app"] # 依存する回答を指定
        interactive: true
      choices:
        microservice:
          - sample-api-1
          - sample-api-2
          - sample-api-3
        deployment:
          - sample-server-1
          - sample-server-2
          - sample-server-3
          - sample-server-4
          - sample-server-5
        job:
          - sample-job-1
          - sample-job-2
          - sample-job-3
          - sample-job-4
          - sample-job-5
    env:
      prompt: "環境名はなんですか？"
      type:
        multiple: true  # 複数選択可能
      choices:
        - dev
        - staging
        - production
    cluster:
      prompt: "クラスターはどこですか？"
      type:
        multiple: true  # 複数選択可能
        dynamic:
          dependency_questions: ["env"] # 依存する回答を指定
      choices:
        dev:
          - dev-cluster-1
          - dev-cluster-2
          - dev-cluster-3
        staging:
          - staging-cluster-1
          - staging-cluster-2
          - staging-cluster-3
        production:
          - production-cluster-1
          - production-cluster-2
          - production-cluster-3
```

**任意のキーに対応した柔軟なスキーマ例:**

```yaml
questions:
  template_question: "something-1"  # どの質問がテンプレート名を決定するかを指定
  order:
    - something-1
    - something-2
    - target-env
  definitions:
    something-1:
      prompt: "Something1?"
      choices:
        - option-a
        - option-b
    something-2:
      prompt: "Something2?"
      type:
        dynamic:
          dependency_questions: ["something-1"]
        interactive: true
      choices:
        option-a:
          - choice-1
          - choice-2
        option-b:
          - choice-3
          - choice-4
    target-env:
      prompt: "対象環境は？"
      type:
        multiple: true
      choices:
        - development
        - staging
        - production
```

**テンプレート質問の指定機能（新機能）:**

従来はヒューリスティック（最初のnon-multiple質問）でテンプレート名を決定していましたが、
`template_question`フィールドで明示的に指定できるようになりました:

```yaml
questions:
  template_question: "appType"  # この質問の回答がテンプレート名になる
  order:
    - region      # 地域選択（multiple可能）
    - appType     # アプリタイプ（テンプレート決定）
    - env         # 環境選択（multiple）
  definitions:
    region:
      prompt: "地域を選択してください"
      type:
        multiple: true
      choices: ["us-east", "us-west", "eu-west"]
    appType:
      prompt: "アプリケーションタイプは？"
      choices: ["web-service", "batch-job", "microservice"]
    env:
      prompt: "環境は？"
      type:
        multiple: true
      choices: ["dev", "staging", "prod"]
```

この場合、`appType`の回答（例: "web-service"）がテンプレート名として使用され、
対応するテンプレートファイル（`.yg/_templates/web-service.yaml`）が読み込まれます。

**下位互換性:**
従来の直接指定形式も引き続きサポートされます:

```yaml
questions:
  app:
    prompt: "アプリの種類は？"
    choices: ["deployment", "job"]
  env:
    prompt: "環境は？"
    type:
      multiple: true
    choices: ["dev", "prod"]
```

`template_question`が指定されていない場合は、従来通りのヒューリスティック
（質問順序で最初のnon-multiple質問）でテンプレート名を決定します。

## テンプレートファイル

### ディレクトリテンプレート（新機能）

./.yg/_templates/microservice/.template-config.yaml

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

./.yg/_templates/microservice/deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Questions.appName}}
  namespace: {{.Questions.env}}
spec:
  replicas: 3
  selector:
    matchLabels:
      app: {{.Questions.appName}}
  template:
    metadata:
      labels:
        app: {{.Questions.appName}}
    spec:
      containers:
      - name: {{.Questions.appName}}
        image: {{.Questions.appName}}:latest
        ports:
        - containerPort: 8080
```

### 単一ファイルテンプレート（従来）

./.yg/_templates/deployment.yaml

```yaml
path: {{.Questions.env}}/{{.Questions.cluster}}/deployment
filename: {{.Questions.appName}}-deployment.yaml
---

appName: {{.Questions.appName}}
env: {{.Questions.env}}
cluster: {{.Questions.cluster}}
```

./.yg/templates/job.yaml

```yaml
path: {{.Questions.env}}/{{.Questions.cluster}}/job
filename: {{.Questions.appName}}-job.yaml
---

appName: {{.Questions.appName}}
env: {{.Questions.env}}
cluster: {{.Questions.cluster}}
```

**任意のキーに対応したテンプレート例:**

./.yg/templates/option-a.yaml

```yaml
path: {{.Questions.target-env}}/{{.Questions.something-2}}
filename: {{.Questions.something-1}}-output.yaml
---

something1: {{.Questions.something-1}}
something2: {{.Questions.something-2}}
targetEnv: {{.Questions.target-env}}
```
