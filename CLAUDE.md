# YAML template generator

このプロジェクトはプロンプトで質問に答えた結果に合わせ、適したYAMLのテンプレートを元にYAMLファイルを出力するGo言語のツールである。

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
- プロンプトの回答は文字列の入力も受け付け、peco (https://github.com/peco/peco) のようにInteractive searchができること。
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
[x] deployment
[] job

# 自由入力を受け付ける。選択肢の中から文字列の前方一致で検索でき、候補からpecoのように矢印キーで選択し、Enterで決定する。
Q. アプリ名は何ですか？
> s
sample-job-1
sample-job-2
sample-job-3
sample-job-4
sample-job-5

> sample-job-1
sample-job-1 # Enterで決定

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

* dev/dev-cluster-1/deployment/sample-job-1-deployment.yaml

appName: sample-job-1
env: dev
cluster: dev-cluster-1

dev/dev-cluster-2/deployment/sample-job-1-deployment.yaml

appName: sample-job-1
env: dev
cluster: dev-cluster-2


dev/dev-cluster-3/deployment/sample-job-1-deployment.yaml

appName: sample-job-1
env: dev
cluster: dev-cluster-3


Q. 出力して問題ないですか? [Y/N]
> Y

# 生成成功の結果が出力される。
generated!
```

`.yg-config.yaml`

```yaml
questions:
  app:
    prompt: "アプリの種類はなんですか？"
    choises:
      - deployment
      - job
  appName:
    prompt: "アプリ名は何ですか？"
    type:
      dynamic:
        dependency_questions: ["app"] # 依存する回答を指定
      interactive: true
    choises:
      app:
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
    choises:
      - dev
      - staging
      - production
  cluster:
    prompt: "クラスターはどこですか？"
    choises:
      dependency_questions: ["app", "env"] # 依存する回答を指定
      app:
        deployment:
          env:
            development:
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
        job:
          env:
            development:
              - dev-cluster
            staging:
              - staging-cluster
            production:
              - production-cluster
```

テンプレートファイル

./.yg/templates/Development.yaml

```yaml
path: {{questions.env}}/{{questions.cluster}}/deployment
filename: {{questions.app}}-deploymentjob.yaml
---

appName: {{quessions.appName}}
env: {{quessions.env}}
cluster: {{quessions.cluster}}
```

./.yg/templates/Job.yaml

```yaml
path: {{questions.env}}/{{questions.cluster}/job/
filename: {{questions.app}}-job.yaml
---

appName: {{quessions.appName}}
env: {{quessions.env}}
cluster: {{quessions.cluster}}
```
