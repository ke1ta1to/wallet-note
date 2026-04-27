# Data Model (DynamoDB)

Single Table Design を採用し、マルチテナント (Organization) でテナント分離する。

## テーブル定義

<!-- prettier-ignore -->
| 項目 | 値 |
|---|---|
| 名前 | `wallet-note-<env>` (例: `wallet-note-development`) |
| 課金モード | PAY_PER_REQUEST |
| PITR | 有効 |
| TTL 属性 | `expiresAt` (Unix epoch 秒) |
| Stream | 必要時のみ有効化 (org 削除カスケード用) |

## キー設計

ベーステーブルは `PK` (HASH) + `SK` (RANGE)、GSI1 は `GSI1PK` + `GSI1SK`。両方とも文字列型。

## エンティティ一覧

```
PK                   SK                          GSI1PK                  GSI1SK
USER#<userId>        ORG#<orgId>                 ORG#<orgId>             USER#<userId>
ORG#<orgId>          META                        ─                       ─
ORG#<orgId>          CATEGORY#<categoryId>       ─                       ─
ORG#<orgId>          TX#<yyyy-mm-dd>#<txId>      ORG#<orgId>#CAT#<catId> TX#<yyyy-mm-dd>
ORG#<orgId>          INVITE#<code>               INVITE#<code>           META
```

ユーザは Cognito 側で識別され、本テーブルに User entity は持たない。`userId` (= Cognito `sub`) は Membership の `PK=USER#<userId>` で間接的に存在する。displayName / email など表示用の属性は Cognito の token から取る。

## アクセスパターン

<!-- prettier-ignore -->
| # | パターン | クエリ |
|---|---|---|
| AP1 | ユーザーから所属 org 一覧 | `Query` Base, `PK=USER#<u>` `begins_with(SK,"ORG#")` |
| AP2 | user × org の権限確認 | `GetItem` Base, `PK=USER#<u>` `SK=ORG#<o>` |
| AP3 | org メタ取得 | `GetItem` Base, `PK=ORG#<o>` `SK=META` |
| AP4 | org のメンバー一覧 | `Query` GSI1, `GSI1PK=ORG#<o>` `begins_with(GSI1SK,"USER#")` |
| AP5 | org のカテゴリ一覧 | `Query` Base, `PK=ORG#<o>` `begins_with(SK,"CATEGORY#")` |
| AP6 | カテゴリ取得 | `GetItem` Base, `PK=ORG#<o>` `SK=CATEGORY#<c>` |
| AP7 | org × 月で取引一覧 | `Query` Base, `PK=ORG#<o>` `begins_with(SK,"TX#2026-04")` |
| AP8 | org × 月 × カテゴリで取引 | `Query` GSI1, `GSI1PK=ORG#<o>#CAT#<c>` `begins_with(GSI1SK,"TX#2026-04")` |
| AP9 | 取引 1件取得 | `GetItem` Base, `PK=ORG#<o>` `SK=TX#<date>#<tx>` |
| AP10 | 招待コード逆引き | `Query` GSI1, `GSI1PK=INVITE#<code>` |

## GSI1 の用途

GSI1 は3用途を兼ねる。`GSI1PK` の prefix がそれぞれ `ORG#<o>` / `ORG#<o>#CAT#<c>` / `INVITE#<code>` で重ならないため、衝突しない。

<!-- prettier-ignore -->
| 用途 | GSI1PK | GSI1SK |
|---|---|---|
| メンバー一覧 | `ORG#<orgId>` | `USER#<userId>` |
| カテゴリ別取引 | `ORG#<orgId>#CAT#<categoryId>` | `TX#<yyyy-mm-dd>` |
| 招待コード逆引き | `INVITE#<code>` | `META` |

## エンティティ属性

Membership:

```json
{
  "PK": "USER#<u>",
  "SK": "ORG#<o>",
  "GSI1PK": "ORG#<o>",
  "GSI1SK": "USER#<u>",
  "type": "Membership",
  "userId": "<u>",
  "orgId": "<o>",
  "role": "owner",
  "joinedAt": "..."
}
```

`role` は `"owner"` または `"member"`。

Organization:

```json
{
  "PK": "ORG#<o>",
  "SK": "META",
  "type": "Organization",
  "orgId": "<o>",
  "name": "...",
  "createdAt": "...",
  "createdBy": "<u>"
}
```

Category:

```json
{
  "PK": "ORG#<o>",
  "SK": "CATEGORY#<c>",
  "type": "Category",
  "orgId": "<o>",
  "categoryId": "<c>",
  "name": "食費",
  "kind": "expense",
  "color": "#ff8800",
  "createdAt": "..."
}
```

`kind` は `"income"` または `"expense"`。カテゴリ自身が収支属性を持つため、取引には `kind` を持たせず、カテゴリ経由で導出する。

Transaction:

```json
{
  "PK": "ORG#<o>",
  "SK": "TX#<yyyy-mm-dd>#<tx>",
  "GSI1PK": "ORG#<o>#CAT#<c>",
  "GSI1SK": "TX#<yyyy-mm-dd>",
  "type": "Transaction",
  "orgId": "<o>",
  "txId": "<tx>",
  "categoryId": "<c>",
  "amount": 1500,
  "date": "2026-04-25",
  "memo": "...",
  "createdAt": "...",
  "createdBy": "<u>"
}
```

`amount` は整数 (円単位)。

`date` は SK と GSI1SK の両方に含まれるため、PATCH で日付を変更する場合は属性更新では済まず、旧 item の delete + 新 item の put を `TransactWriteItems` でアトミックに行う。

Invite:

```json
{
  "PK": "ORG#<o>",
  "SK": "INVITE#<code>",
  "GSI1PK": "INVITE#<code>",
  "GSI1SK": "META",
  "type": "Invite",
  "code": "<code>",
  "orgId": "<o>",
  "role": "member",
  "invitedBy": "<u>",
  "expiresAt": 1714492800,
  "createdAt": "..."
}
```

`expiresAt` は DynamoDB TTL 属性だが、TTL の削除は最大 48 時間遅れる。招待検証 (`GET /invites/{code}`、`POST /invites/{code}/accept`) は必ずアプリ側で `expiresAt > now` を確認する。TTL はストレージ掃除のみと見なす。

## ID 採番

`userId` は Cognito の `sub` をそのまま使う。`orgId` / `categoryId` / `txId` は **UUIDv7** (RFC 9562、時系列 sortable)。招待 `code` は6文字英数字 (大文字、紛らわしい字 0/O/1/I/L を除外、約30文字種)、生成時に `ConditionExpression` で重複チェックしてリトライする。

## マルチテナント分離

ほぼ全データが `PK=ORG#<orgId>` で物理分離される。Membership だけは AP1 を効率的に引くために `PK=USER#<userId>` 側に持たせる。

## 認可

org スコープの API はすべて AP2 (`GetItem(USER#<u>, ORG#<o>)`) を最初に通し、Membership の有無と role を判定する。

## 主要操作

### 組織作成 (`POST /orgs`)

`ORG#<o>/META` (Organization) と `USER#<u>/ORG#<o>` (Membership, role=owner) を `TransactWriteItems` でアトミックに作成する。Org 単独で `PutItem` すると、Membership 書き込み失敗時に owner の居ない孤児 Org が残るため。

### 所属 org 一覧 (`GET /me/orgs`)

AP1 で Membership 一覧を引いたあと、得られた orgId 群を `BatchGetItem` で `ORG#<o>/META` から取得し、orgName と join して返す。Membership に orgName を冗長保存しないため、Org 名変更時の cascade 更新は不要。

## 集計

MVP では取引を取得して Go 側で都度集計する。将来データ量が増えたら、Streams で `ORG#<o>/SUMMARY#<yyyy-mm>` を持つ事前集計テーブルに昇格する。

## カスケード削除

org 削除時は Streams + Lambda で `PK=ORG#<o>` の全件と GSI1 経由の Membership を削除する。
