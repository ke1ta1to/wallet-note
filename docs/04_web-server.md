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
github.com/google/uuid
github.com/go-playground/validator/v10
```

HTTP 周りは標準 `net/http` のみ。

## ディレクトリ構造

機能別コロケーションで切る。各機能パッケージ内に handler / repository / dynamo / model / dto を同居させる。

```
web-server/
├── cmd/wallet-note/main.go        ListenAndServe(":8080") を呼ぶだけ。配線は server.NewMux に委譲
├── internal/
│   ├── app/                       repos / service / middleware / handlers の配線 (main + tests 共通)
│   ├── transaction/               取引機能
│   ├── category/                  カテゴリ機能
│   ├── organization/              Org + Membership (同集約、メンバー管理は member_handler.go)
│   ├── invite/                    招待機能 (service.go で TransactWriteItems)
│   ├── user/                      /me, /me/orgs (organization 経由で Membership 取得)
│   ├── auth/                      JWT claims 抽出、WithOrgAuth ミドルウェア
│   │                              MembershipReader interface を内部宣言する
│   ├── testutils/                 request test 用 helper (DDB Local, Setup)
│   └── shared/                    横断基盤 (誰にも依存しない)
│       ├── apperror/              ErrNotFound, ErrForbidden 等
│       ├── httpx/                 JSON ヘルパ、エラー → HTTP 変換
│       ├── idgen/                 UUIDv7 と招待コードの生成
│       ├── ddb/                   DynamoDB クライアント生成
│       ├── router/                Registrar interface
│       └── config/                環境変数読み込み
└── go.mod
```

`internal/app.NewMux(db, tableName)` が application 全体の wire-up を担う。`main.go` も request test (`testutils.Setup`) も同じ関数を呼ぶことで、本番と test の wiring drift が起こらない。

各機能パッケージは概ね次のファイル構成を取る。

```
handler.go       HTTP ハンドラ
handler_test.go  request test (実 DDB に当てて HTTP I/O + DDB 状態を検証)
repository.go    Repository インターフェース
dynamo.go        DynamoDB 実装
model.go         ドメインモデル
dto.go           API DTO + 変換 (ドメインモデルとは分離)
service.go       必要時のみ (複数リポジトリにまたがる処理用)
```

## 依存ルール

`shared` は誰にも依存しない (基盤層)。`auth` は機能パッケージに直接依存せず、必要なら自分の中で interface (例: `MembershipReader`) を宣言し、main で organization の実装を渡してもらう (Go の implicit interface)。

`organization` は他機能から参照される側で、`OrgRepository` と `MembershipRepository` を提供する。`user` は所属 org 一覧の取得で `organization` に依存する。

## 設計上の決定

Repository はエンティティ単位で切る。Service 層は最初から作らず、招待受諾のように複数リポジトリにまたがる処理が出たタイミングで追加する。

DTO はドメインモデルと分離する。内部表現と API のレスポンス形式を独立に進化させたいため。

## レイヤリング規約

機能パッケージ (`organization`, `transaction`, …) は次の責務で構成する:

- **handler**: HTTP I/O のみ。リクエスト解析 → repository / service 呼び出し → レスポンス書き込み
- **repository (interface)**: 1 エンティティの永続化契約。`feature/repository.go` で interface を宣言、`dynamo.go` などで実装
- **service**: 複数 repository をまたぐ処理 (`TransactWriteItems`、cross-entity 整合性)。**最初から作らず、必要が出たタイミングで追加**

handler は repository (interface) を直接持ち、CRUD はそれだけで完結する。複数 repository を跨ぐ処理 (例: `POST /orgs` の Org + Membership atomic 作成) が出たときに service.go を導入する。

repository は interface (storage 差し替えの余地、テストでの mock 注入のため)、service は struct (1 impl 想定、抽象化の必要が出たら interface に格上げ)。

## テスト戦略

**Rails の request test 主体方針** を Go に持ち込む。Handler から full stack を通して実 DDB に当てるテストを default にし、内部 layer (service / repository) の細かい unit test は基本書かない。

| レイヤ | テスト種別 | 手段 |
|---|---|---|
| handler | request test | `httptest.NewRecorder` + 実 Repository / Service / Middleware を本物のまま wire、DDB Local に書き込み・読み出しする。HTTP I/O + DDB 状態の両方を検証 |
| service | (request test 経由でカバー) | 単独 unit test は基本書かない。複雑な business logic が出てきた時だけ追加 |
| repository (DDB impl) | (request test 経由でカバー) | 単独 unit test は基本書かない。アルゴリズム的に重い箇所が出てきた時だけ追加 |
| middleware (auth) | (request test 経由でカバー) | 401 / 403 / 200 の境界は handler の request test に統合される |
| 例外 (純粋アルゴリズム) | unit | 招待コード生成のリトライ等、外部 I/O に依存しない関数は package-local の unit test を書く |
| e2e | 自動 (CI/CD で curl) | CloudFront → API GW → Lambda の経路を主要 endpoint で 1 リクエストずつ |

### テスト基盤

- **DDB Local を docker compose で起動**。`amazon/dynamodb-local:latest` を `docker-compose.yml` で定義、`make test` が `docker compose up -d dynamodb-local` を経由してから `go test` を呼ぶ
- Test 用 table 名は `wallet-note-test` (固定)。Schema 定義は `internal/testutils/ddb.go` の `createTable` 関数に集約 (Terraform 側 schema と手動同期)
- Test 間の隔離は **各 test の冒頭で全 item を delete** (`testutils.ResetTable`)。table は使い回し
- testcontainers-go は採用しない。docker compose で十分シンプルで、起動忘れは Makefile target で潰せる

### Mock 戦略

Repository / Service は **mock せず実物を使う**。Mock 手書きや `mockery` / `gomock` のコード生成は不要。Request test で full stack を通す方針なら、Repository / Service の内部仕様変更は HTTP 入出力 + DDB 状態の合意でしか検証されない (= refactor 耐性が高い)。

ペイロードは snake_case。ID はすべて UUIDv7 文字列、日付は `YYYY-MM-DD`、タイムスタンプは ISO8601。バリデーションは `go-playground/validator/v10` の struct tag で書く。OpenAPI の制約 (required, pattern 等) と struct tag は手動で同期する。本格化したら `oapi-codegen` での自動化を検討する。

## レスポンス形状

リスト系エンドポイント (`GET /me/orgs`, `GET /orgs/{orgId}/transactions` 等) は次の形に統一する:

```json
{
  "items": [ ... ],
  "nextCursor": "..."
}
```

`nextCursor` は optional (省略時は次ページ無し)。値は DDB の `LastEvaluatedKey` を base64 化したもの。`limit/offset` ベースは採用しない (大規模 query で重い、DDB の cursor 設計と齟齬)。

エラーは:

```json
{ "message": "..." }
```

ステータスコードは HTTP 標準 (400 invalid input、401 unauthorized、403 forbidden、404 not found、409 conflict、500 internal)。

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
