# YAML template generator

このプロジェクトはプロンプトで質問に答えた結果に合わせ、適したYAMLのテンプレートを元にYAMLファイルを出力するGo言語のツールである。

## Claude Rule

- 日本語で回答して
- 修正後にREADME.md/CLAUDE.mdを修正して
- 修正後にテストと凛とを実行して成功するか確認して、失敗したら修正して。
- 最後にGitコミットを実行して、GitHubにpushして。

## Guidelines

- Style: Go言語の標準プロジェクト構成で開発する。
- Coding: 可読性の高いコードを心がける。
  - 参考: https://google.github.io/styleguide/go/guide
- Documentation: godocで出力可能なコメントを作成する。
- Test: テストカバレッジ90%以上のテストコードを作成する。テストケースをテストディレクトリにMarkdownでまとめる。
- CI/CD: GitHub ActuonsでCIによるテスト、CDによるgo releaserを実行し、GitHub上でバイナリの配布も行う。go installでインストール可能とする。
- Quality: - エラーハンドリングを適切に行うこと。異常系のテストケースを必ず作成し、問題があった場合は不適切な出力は実行しないようにする。
- CommitMessage: コミットメッセージはSemantic Commitで、mainブランチへのマージでセマンティックバージョンを上げる。

## Tool Design

目的は、環境ディレクトリ、クラスターのディレクトリに応じて、アプリの系統に合わせたテンプレートを元にYAMLファイルを出力すること。
ユーザーはプロンプトに答えるだけで、テンプレートに合わせたファイルを、指定ディレクトリに出力可能となる。

ポイントは以下である。

- プロンプトに応じて、出力ディレクトリを指定できること。
- プロンプトに応じて、利用するテンプレートファイルを選択できること。
- プロンプトは前の回答に応じて、動的に選択可能なこと。
- プロンプトの回答は文字列の入力も受け付け、リアルタイム検索でInteractive searchができること。
- Interactive search機能では、文字を入力すると選択肢がリアルタイムでフィルタリングされ、全文一致・部分一致の両方に対応。
- テンプレートファイルに、プロンプトで受け取った値を指定のYAML Pathに合わせて値を埋められること。
- 出力ディレクトリの質問プロンプトでは複数選択可能であること。
- プロンプトの質問回答は、CLIオプションでも設定可能なこと。
- プロンプト中はCtrl + Cで解除可能なこと。強制終了のシグナルを受け取った場合は元に戻すこと。
- キャッチーな見た目のTerminal UIを提供すること。

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

# 生成成功の結果が出力される。
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
