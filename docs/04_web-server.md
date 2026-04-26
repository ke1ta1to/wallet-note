# web-server

Go の Lambdalith。AWS Lambda Web Adapter (LWA) を Layer として乗せ、アプリは普通の HTTP サーバとして書く。

## ランタイム構成

<!-- prettier-ignore -->
| 項目 | 値 |
|---|---|
| ランタイム | `provided.al2023` (custom runtime) |
| アーキテクチャ | arm64 (Graviton、約20%安い) |
| Layer | AWS Lambda Web Adapter (公式 ARN を attach) |
| ポート | 8080 (LWA から HTTP で受ける) |

アプリのエントリは `net/http.ListenAndServe(":8080", mux)`。Go 1.22+ の `net/http.ServeMux` で `POST /orgs/{orgId}/transactions` のような path patterns を直接書ける。chi 等のルータも `aws-lambda-go` も依存に含めない。

## 採用ライブラリ

```
github.com/aws/aws-sdk-go-v2 (config, service/dynamodb,
                              feature/dynamodb/attributevalue,
                              service/dynamodb/expression)
github.com/oklog/ulid/v2
github.com/go-playground/validator/v10
```

HTTP 周りは標準 `net/http` のみ。

## ディレクトリ構造

機能別コロケーションで切る。各機能パッケージ内に handler / repository / dynamo / model / dto を同居させる。

```
web-server/
├── cmd/server/main.go             DI 組み立て + ListenAndServe(":8080")
├── internal/
│   ├── transaction/               取引機能
│   ├── category/                  カテゴリ機能
│   ├── organization/              Org + Membership (同集約、メンバー管理は member_handler.go)
│   ├── invite/                    招待機能 (service.go で TransactWriteItems)
│   ├── user/                      /me, /me/orgs (organization 経由で Membership 取得)
│   ├── auth/                      JWT claims 抽出、WithOrgAuth ミドルウェア
│   │                              MembershipReader interface を内部宣言する
│   └── shared/                    横断基盤 (誰にも依存しない)
│       ├── apperror/              ErrNotFound, ErrForbidden 等
│       ├── httpx/                 JSON ヘルパ、エラー → HTTP 変換
│       ├── idgen/                 ULID と招待コードの生成
│       ├── ddb/                   DynamoDB クライアント生成
│       ├── router/                ServeMux 構築
│       └── config/                環境変数読み込み
└── go.mod
```

各機能パッケージは概ね次のファイル構成を取る。

```
handler.go     HTTP ハンドラ
repository.go  Repository インターフェース
dynamo.go      DynamoDB 実装
model.go       ドメインモデル
dto.go         API DTO + 変換 (ドメインモデルとは分離)
service.go     必要時のみ (複数リポジトリにまたがる処理用)
```

## 依存ルール

`shared` は誰にも依存しない (基盤層)。`auth` は機能パッケージに直接依存せず、必要なら自分の中で interface (例: `MembershipReader`) を宣言し、main で organization の実装を渡してもらう (Go の implicit interface)。

`organization` は他機能から参照される側で、`OrgRepository` と `MembershipRepository` を提供する。`user` は所属 org 一覧の取得で `organization` に依存する。

## 設計上の決定

Repository はエンティティ単位で切る。Service 層は最初から作らず、招待受諾のように複数リポジトリにまたがる処理が出たタイミングで追加する。

DTO はドメインモデルと分離する。内部表現と API のレスポンス形式を独立に進化させたいため。

ペイロードは camelCase。ID はすべて ULID 文字列、日付は `YYYY-MM-DD`、タイムスタンプは ISO8601。バリデーションは `go-playground/validator/v10` の struct tag で書く。OpenAPI の制約 (required, pattern 等) と struct tag は手動で同期する。本格化したら `oapi-codegen` での自動化を検討する。

## API エンドポイント

```
GET    /me
GET    /me/orgs

POST   /orgs
GET    /orgs/{orgId}
PATCH  /orgs/{orgId}                                [owner]
DELETE /orgs/{orgId}                                [owner]
GET    /orgs/{orgId}/members
DELETE /orgs/{orgId}/members/{userId}               [owner]

POST   /orgs/{orgId}/invites                        [owner]
GET    /orgs/{orgId}/invites                        [owner]
DELETE /orgs/{orgId}/invites/{code}                 [owner]
GET    /invites/{code}                              (ログインのみ要求)
POST   /invites/{code}/accept                       (ログインのみ要求)

POST   /orgs/{orgId}/categories
GET    /orgs/{orgId}/categories
GET    /orgs/{orgId}/categories/{categoryId}
PATCH  /orgs/{orgId}/categories/{categoryId}
DELETE /orgs/{orgId}/categories/{categoryId}

POST   /orgs/{orgId}/transactions
GET    /orgs/{orgId}/transactions?from=&to=&categoryId=&cursor=&limit=
GET    /orgs/{orgId}/transactions/{txId}
PATCH  /orgs/{orgId}/transactions/{txId}
DELETE /orgs/{orgId}/transactions/{txId}

GET    /orgs/{orgId}/summary?month=2026-04
```

`[owner]` 印が付いたエンドポイントは owner 限定。それ以外は member 以上で叩ける。

## ページネーション

取引一覧は cursor ベース。DynamoDB の `LastEvaluatedKey` を base64 エンコードして `nextCursor` として返す。
