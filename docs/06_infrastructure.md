# infrastructure (Terraform)

## State 管理

Terraform 1.10+ の S3 native lock (`use_lockfile = true`) を使う。DynamoDB lock テーブルは作らない。

State 用の S3 バケットは PoC アカウントの既存バケット (`tfstate-075472845547-ap-northeast-1-an`) を流用するため、bootstrap モジュールは作らない。

backend 設定:

```hcl
terraform {
  backend "s3" {
    bucket       = "tfstate-075472845547-ap-northeast-1-an"
    key          = "wallet-note/<env>/terraform.tfstate"
    region       = "ap-northeast-1"
    use_lockfile = true
    encrypt      = true
  }
}
```

`<env>` は `development` or `production`。

## ディレクトリ構造

```
infrastructure/
├── modules/
│   ├── cognito/     User Pool + Client + Managed Login UI Domain + Branding
│   ├── dynamodb/    DynamoDB Table (PK/SK + GSI1, TTL, PITR, on-demand)
│   ├── web-server/  Lambda (LWA Layer, arm64) + API GW HTTP API + JWT Authorizer + IAM Role
│   └── web-client/  S3 (no public, OAC) + CloudFront (S3 + API GW origins)
└── environments/
    ├── development/
    │   ├── main.tf       modules を組み合わせる
    │   ├── locals.tf     name_prefix などの派生値
    │   ├── variables.tf
    │   ├── outputs.tf
    │   ├── backend.tf    S3 backend 設定
    │   ├── providers.tf  AWS provider (ap-northeast-1)
    │   └── terraform.tfvars
    └── production/
        └── ... (development と同構造)
```

新しい環境を追加するときは `environments/<env>/` をコピーして tfvars だけ書き換える。

## 命名規約

`locals.name_prefix = "${var.project}-${var.environment}"` を組み立て、各 module に `name_prefix` として渡す。各 module はその prefix から個別リソース名を生成する (例: `${var.name_prefix}-spa`)。

これにより全リソースが `<project>-<env>-...` で揃い、AWS Console での視認性と環境追加のしやすさを両立する。

## ドメイン

独自ドメインは使わず、CloudFront と Cognito の発行ドメインで運用する。これにより ACM、Route 53、us-east-1 provider alias が一切不要になる。

Cognito の Domain Prefix は Terraform で確保する (グローバルユニーク)。Cognito の Callback URL は `module.web_client.distribution_domain` を参照して組み立てるため、tfvars 更新や2段階 apply は不要 (環境内で web-client → cognito の依存が解決される)。

## リージョン

全リソース ap-northeast-1 (東京)。

## CloudFront 構成

Distribution は2つの Origin (S3 と API Gateway) を持つ。

```
CloudFront Distribution
├─ Origin 1: S3 (web-client, OAC)
├─ Origin 2: API Gateway HTTP API
│
├─ Default Behavior (* → S3)
│    ├─ Viewer Protocol: HTTPS only
│    └─ Cache: hashed assets max-age=31536000、index.html no-cache
│
└─ Behavior /api/* → API Gateway Origin
     ├─ Cache: Disabled
     ├─ Origin Request Policy: AllViewerExceptHostHeader (公式 docs が API GW origin にはこれを推奨)
     └─ Forward Authorization header (重要)
```

`Authorization` ヘッダの転送漏れは詰みポイントなので注意する。

## SPA fallback

`custom_error_response` (distribution 全体の 403/404 → `/index.html`) は採用しない。distribution-level で適用されるため `/api/*` の 403/404 まで `/index.html` に書き換えてしまい、API レスポンスが壊れる。

代わりに default behavior に CloudFront Function (viewer-request) を attach し、拡張子を含まないパスを `/index.html` に rewrite する。`/api/*` は別の ordered cache behavior に分離されているため Function は発火しない (Function association は behavior 単位)。Function コードは `infrastructure/modules/web-client/spa-fallback.js` (runtime `cloudfront-js-2.0`)。

判定は `request.uri.includes('.')` による単純な「ドット有無」のみ。アセットは `/assets/<hash>.js` 等で必ず拡張子を持ち、SPA route には拡張子を含むセグメントを設けない前提。

## Cognito Callback URL

`aws_cognito_user_pool_client.callback_urls` は environment 側 (`environments/<env>/main.tf`) で組み立てて module に渡す。CloudFront ドメインは `module.web_client.distribution_domain` を参照する。development では追加で localhost も入れる。

```
- https://<cloudfront-domain>/auth/callback   各環境の本物 (module 参照)
- http://localhost:5173/auth/callback         development のみ追加
```

## Lambda 構成

`provided.al2023` の custom runtime、arm64。AWS Lambda Web Adapter の公式 Layer (リージョン × アーキテクチャで ARN が異なる) を attach する。

環境変数:

```
WALLET_NOTE_TABLE=wallet-note-<env>
PORT=8080                          # LWA と Go app が同じ env を読む
AWS_LWA_REMOVE_BASE_PATH=/api      # /api/* prefix を Go に渡す前に剥がす
```

`provided.al2023` では Lambda が `bootstrap` を直接 exec し、LWA は Lambda Extension として動くため `AWS_LAMBDA_EXEC_WRAPPER` は不要 (これは managed runtime + zip の場合に使う変数)。

`PORT` は LWA の `AWS_LWA_PORT` の fallback (公式 gin-zip サンプルもこの形式)。1つの env var で LWA forward 先と Go listen port をペアリングし、ずれる余地を消す。

IAM Role は DynamoDB CRUD と CloudWatch Logs に絞る。

## デプロイ責務の分離

Terraform は AWS リソースの「箱」だけ定義する。Lambda コードのバイナリと Web Client のビルド成果物は別ルートでデプロイする。

GitHub Actions のワークフロー側に直接デプロイコマンドを書き、ローカル動作確認用には `scripts/` 配下にシェルスクリプトを置く。両者は意図的に共通化せず、それぞれシンプルに保つ。詳細は 07_cicd.md と 08_local-dev.md を参照。
