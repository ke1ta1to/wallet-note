# Local Development

## 基本方針

dev 環境の AWS リソースを直接使う。DynamoDB Local や Cognito Local は使わない。Cognito Local では Managed Login UI が動かないため、本物との挙動差で詰まりやすいのが理由。

<!-- prettier-ignore -->
| コンポーネント | 接続先 |
|---|---|
| DynamoDB | dev 環境の本物 (`wallet-note-development`) |
| Cognito | dev 環境の本物 (Managed Login UI 含む) |
| Lambda Go アプリ | ローカルで `go run ./cmd/wallet-note` (LWA は経由せず普通の HTTP サーバ) |
| Web Client | Vite dev server |

## 認証検証 (Web Server コードに分岐を入れない)

「`LOCAL_DEV=true` なら認証バイパス」のような環境変数による分岐を Web Server に入れない。環境変数の設定ミスで本番環境でも認証スキップになる穴を作りたくないため。

代わりに、Vite proxy 層で `x-amzn-request-context` ヘッダを fake 付与する。Web Server は本番と完全に同じコード (`x-amzn-request-context` から claims を取り出すだけ) のまま動かせる。Vite dev server は本番では使われないので、この捏造は本番に一切影響しない。

JWT 自体は本物の Cognito から取ったものを使うため、payload の中身は信頼できる。署名検証はローカルでは省略する (本番側は API Gateway が検証している)。

vite.config.ts の proxy 設定例:

```ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/api/, ''),
      configure: (proxy) => {
        proxy.on('proxyReq', (proxyReq, req) => {
          const auth = (req.headers.authorization || '').replace('Bearer ', '')
          if (!auth) return
          const payload = JSON.parse(
            Buffer.from(auth.split('.')[1], 'base64url').toString()
          )
          const requestContext = {
            authorizer: { jwt: { claims: payload } },
          }
          proxyReq.setHeader('x-amzn-request-context', JSON.stringify(requestContext))
        })
      },
    },
  },
},
```

## ポート

<!-- prettier-ignore -->
| サービス | ポート |
|---|---|
| Vite dev server | 5173 |
| Go HTTP サーバ | 8080 (本番の LWA と同じ) |

Vite proxy で `/api/*` を `http://localhost:8080` に流す。

## 環境変数

Web Server:

```
WALLET_NOTE_TABLE=wallet-note-development
AWS_REGION=ap-northeast-1
```

AWS 認証情報は `~/.aws/credentials` から取得する。

フロント (`web-client/.env.local`、gitignore):

```
VITE_API_BASE=/api
VITE_COGNITO_USER_POOL_ID=ap-northeast-1_xxxxx
VITE_COGNITO_CLIENT_ID=xxxxx
VITE_COGNITO_DOMAIN=wallet-note-dev-xxxxx.auth.ap-northeast-1.amazoncognito.com
```

値はすべて Terraform output から取る。

## タスクランナー

ルート Makefile と web-client の最小 pnpm scripts の2層に分ける。

ルート Makefile はモノレポ全体の入口で、ローカル開発者が触るコマンドはここに集約する。

<!-- prettier-ignore -->
| ターゲット | 内容 |
|---|---|
| `make dev` | dev-server と dev-client を並列起動 |
| `make dev-server` | Go HTTP サーバ起動 |
| `make dev-client` | Vite dev server 起動 |
| `make gen-api` | OpenAPI から型生成 |
| `make test` | go test |
| `make lint` | go vet + web-client lint |
| `make deploy-server-dev` | Lambda Go ビルド + update-function-code (development) |
| `make deploy-client-dev` | Web Client build + S3 sync + CloudFront invalidate (development) |
| `make tf-plan-dev` | development の terraform plan |
| `make tf-apply-dev` | development の terraform apply |
| `make tf-plan-prod` | production の terraform plan (apply は GitHub release 経由のみ) |

web-client/package.json の scripts は Vite が直接必要とする最小限だけ持つ。

```
dev            vite
build          tsc + vite build
preview        vite preview
generate:api   openapi-typescript ../openapi/openapi.yaml -o ./src/api/schema.d.ts
lint           tsc --noEmit
```

CI (GitHub Actions) はこの pnpm scripts を直接叩く。Makefile を経由させない。

## scripts/

ローカルから手動デプロイする用のシェルスクリプトを置く。

```
scripts/deploy-server-dev.sh   Lambda Go ビルド + update-function-code
scripts/deploy-client-dev.sh   Web Client build + S3 sync + CloudFront invalidate
```

production 用は基本作らない。production へのデプロイは GitHub Release 経由が原則。

## Cognito Callback URL

dev の User Pool Client には本番ドメインとローカル URL の両方を登録する。

```
https://<cloudfront-domain>/auth/callback     本物
http://localhost:5173/auth/callback           ローカル開発用
```

これは Terraform の `aws_cognito_user_pool_client.callback_urls` に配列で渡す。

## 起動フロー

```bash
# 1. AWS CLI セットアップ (PoC アカウント) ※初回のみ
aws configure --profile wallet-note-poc

# 2. development 環境を Terraform で構築 ※初回のみ
make tf-apply-dev

# 3. .env.local 作成 (Terraform output から値を取る) ※初回のみ

# 4. ローカル起動
make dev

# → http://localhost:5173 を開く
# → Cognito Managed Login UI でログイン
# → コールバック → Vite が JWT を Authorization ヘッダで API へ
# → Vite proxy が fake context 付与 → ローカル Go アプリが処理
# → AWS dev DynamoDB に書き込み
```
