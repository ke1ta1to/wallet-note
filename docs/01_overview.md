# Overview

個人用家計簿の SaaS。マルチテナント (Organization) を採用し、1つの Cognito アカウントで複数 org を切り替えられる。招待コードで org に他ユーザーを追加できる。

## 技術スタック

<!-- prettier-ignore -->
| 領域 | 採用 |
|---|---|
| Frontend | React 19 + Vite + TypeScript + TanStack Router + Mantine |
| Backend | Go (Lambdalith) + AWS Lambda Web Adapter (LWA, arm64) |
| 認証 | Amazon Cognito (Managed Login UI) |
| API | API Gateway HTTP API + JWT Authorizer |
| データ | DynamoDB (Single Table, on-demand, PITR 有効) |
| 配信 | S3 + CloudFront + OAC |
| IaC | Terraform (S3 native lock) |
| CI/CD | GitHub Actions (OIDC) |
| リージョン | ap-northeast-1 |
| AWS アカウント | PoC 用 (075472845547) |

## モノレポ構成

```
wallet-note/
├── openapi/              OpenAPI SOT (手書き)
├── web-client/           Frontend
├── web-server/           Backend (Go)
├── infrastructure/       Terraform
├── docs/                 設計ドキュメント (本ディレクトリ)
└── scripts/              ローカル/開発用シェルスクリプト
```

## データフロー

本番:

```
Browser
  → CloudFront (default → S3、/api/* → API GW)
  → API Gateway HTTP API (JWT Authorizer で Cognito Access Token を検証)
  → Lambda (Go + LWA, arm64)
      → DynamoDB
```

ローカル開発:

```
Browser
  → Vite dev server (proxy で /api/* → localhost:8080)
  → Go HTTP server (LWA なしで直接起動)
      → AWS dev DynamoDB
```

## ドメイン

独自ドメインは取得済みだが、当面は CloudFront / Cognito の発行ドメインを使う。ACM・Route53 は不要。

## 環境

GitHub Environment 名は `Production` / `Development` (先頭大文字)、Terraform および AWS リソース名は `production` / `development` (小文字)。両者は CI 上で `matrix.include` でペアリングする。

`Development` は main にマージしたタイミングで自動デプロイ。`Production` は GitHub Release を published した時に手動承認を経てデプロイする。

## ドキュメント索引

1. [01_overview.md](./01_overview.md) — 本ファイル
2. [02_data-model.md](./02_data-model.md) — DynamoDB 設計
3. [03_auth.md](./03_auth.md) — Cognito + API Gateway
4. [04_web-server.md](./04_web-server.md) — Go Lambda コード構造
5. [05_web-client.md](./05_web-client.md) — Frontend
6. [06_infrastructure.md](./06_infrastructure.md) — Terraform
7. [07_cicd.md](./07_cicd.md) — GitHub Actions
8. [08_local-dev.md](./08_local-dev.md) — ローカル開発
