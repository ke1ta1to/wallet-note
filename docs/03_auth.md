# Auth

Cognito User Pool は「ユーザーの識別」だけを担当し、Organization とメンバーシップはアプリ側 (DynamoDB) で完全に管理する。

## Cognito

<!-- prettier-ignore -->
| 項目 | 設定 |
|---|---|
| User Pool | シングル (全ユーザーが1プールに乗る) |
| 認証 UI | Managed Login UI (旧 Hosted UI の刷新版) |
| 認証手段 | メール / パスワードのみ |
| MFA | なし (将来追加余地は残す) |
| App Client | Public、Client Secret なし、Authorization Code Flow + PKCE |
| Domain | Terraform で Domain Prefix を確保 (グローバルユニーク) |

## トークン

API リクエストの `Authorization: Bearer <token>` には Cognito の Access Token を載せる。ID Token はクライアント (SPA) 内でユーザー表示用にだけ使う。

API Gateway HTTP API の JWT Authorizer が Access Token を自動検証する。`audience` は Cognito App Client ID、`issuer` は `https://cognito-idp.ap-northeast-1.amazonaws.com/<userPoolId>`。

検証済みの claims は `requestContext.authorizer.jwt.claims` に入って Lambda に渡る。LWA を経由する場合は `x-amzn-request-context` HTTP ヘッダに JSON で乗せられる。Go 側で JWT を自前検証する必要はない。

## API Gateway

REST API ではなく HTTP API を使う。コストが約1/3、レイテンシも低い。Lambdalith なので `$default` ルートで全リクエストを単一 Lambda に流し、ルーティングは Lambda 側で行う。

## URL 構造

`orgId` はパスパラメータで渡す (`/orgs/{orgId}/...`)。RESTful に表現でき、ログ・トレースでも org が一目で分かるため。ヘッダや JWT 埋め込みは採用しない。

## 認可ミドルウェア (Lambda Go)

org スコープの API はすべて以下を通る。

1. `x-amzn-request-context` から `authorizer.jwt.claims.sub` を取り出して `userId` を得る
2. URL のパスパラメータから `orgId` を得る
3. `GetItem(USER#<userId>, SK=ORG#<orgId>)` で Membership を引く
4. 取れなければ 403、取れたら role を見て要求権限と照合する

## 招待フロー

SES は使わない。招待コードは6文字英数字で、有効期限7日、TTL で自動削除する (詳細は data-model.md の Invite を参照)。

MVP では owner が招待コードを生成し、画面にコードと URL を表示する。owner はそれをコピーして手動で (口頭・LINE・メール等で) 共有する。

受諾は `https://<host>/invite?code=<code>` のリンクをフロントが受けて、入力欄にコードを prefill する。ログイン後に `POST /invites/{code}/accept` を叩くと、`TransactWriteItems` で Membership 作成と Invite 削除が原子的に行われる。

将来的にメール送信機能が必要になったら Resend を採用する想定。DNS 認証だけで済み本番審査がないため、SES より個人プロジェクト向き。

## レート制御

`/invites/{code}/accept` は API Gateway throttling で IP あたり 10 req/min 程度に制限し、ブルートフォースを防ぐ。
