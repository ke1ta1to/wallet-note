# CI/CD (GitHub Actions)

## ファイル構成

```
.github/workflows/
├── infrastructure.yml         Terraform plan / apply
├── web-server.yml             Lambda Go build + deploy
├── web-client.yml             Web Client build + deploy
├── openapi.yml                PR で schema.d.ts 差分チェックのみ
├── _terraform-apply.yml       Reusable
├── _lambda-deploy.yml         Reusable
└── _web-client-deploy.yml    Reusable
```

## 命名と環境

GitHub Environment は表示用に先頭大文字 (`Production` / `Development`)、Terraform および AWS リソース名は小文字 (`production` / `development`) を使う。両者は workflow 内の `matrix.include` で `github_env` と `env` をペアリングして変換する。

## トリガーポリシー

<!-- prettier-ignore -->
| イベント | 動作 |
|---|---|
| Pull Request | plan / test / build / openapi 差分チェック (両環境 plan は matrix で並列) |
| push to main | `Development` の apply / deploy が全 workflow で起動 |
| release published | `Production` の apply / deploy が全 workflow で起動 (Environment で手動承認) |
| workflow_dispatch | `Development` / `Production` を選んで手動実行 |

`paths` フィルタは使わない。フィルタの漏れで「変更したのにデプロイされない」事故を防ぐため、main マージや release 時は全 workflow を素直に走らせる。

## 構造方針

plan / test / build は PR で両環境並列 (matrix) に走らせる、または環境非依存ならシングルジョブ。

apply / deploy は環境ごとに別ジョブに分ける。trigger が `push` と `release` で異なるため matrix が合わない。中身は Reusable Workflow に切り出して重複を抑える。

## AWS 認証

GitHub Actions OIDC で IAM Role を AssumeRole する。Access Key は使わず、リポジトリ Secrets はほぼ不要になる。IAM OIDC Provider (`token.actions.githubusercontent.com`) は Terraform で1度だけ作成する。

## IAM Role 命名規則

```
github-actions-wallet-note-{component}-{env}

例:
  github-actions-wallet-note-terraform-development
  github-actions-wallet-note-terraform-production
  github-actions-wallet-note-api-development
  github-actions-wallet-note-api-production
  github-actions-wallet-note-web-development
  github-actions-wallet-note-web-production
```

`component` は `terraform` / `api` / `web` の3種。

## GitHub Environment 設定

リポジトリの Settings から手動で構成する。

```
Production
├── Required reviewers: 自分自身       (手動承認ゲート)
└── Variables
    └── CLOUDFRONT_DISTRIBUTION_ID = E1XXXXXXXXX

Development
├── (Required reviewers なし、自動)
└── Variables
    └── CLOUDFRONT_DISTRIBUTION_ID = E2YYYYYYYYY
```

リポジトリは public なので Required reviewers が無料で使える。`CLOUDFRONT_DISTRIBUTION_ID` は当面手動で入れ、将来的に SSM Parameter Store 経由の自動化を検討する。

## OpenAPI workflow の特殊性

PR でのみ実行する。`pnpm run generate:api` を走らせて `web-client/src/api/schema.d.ts` に差分が出たら失敗させる。これにより「openapi.yaml は変更したのに schema.d.ts がコミットされていない」状態を防ぐ。main / release では走らせない (デプロイ workflow が build 中に generate するため)。

## 値の伝達

`CLOUDFRONT_DISTRIBUTION_ID` は GitHub Environment の Variables に手動で書く。Lambda 関数名と S3 バケット名は規則ベースで決まる (`wallet-note-<env>-api`、`wallet-note-<env>-web-client-<account_id>`)。
