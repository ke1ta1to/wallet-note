# web-client

React 19 + Vite + TanStack Router + Mantine。スマホで触ることを前提とした PWA。

## 採用ライブラリ

<!-- prettier-ignore -->
| 用途 | ライブラリ |
|---|---|
| Framework | React 19 |
| Build | Vite |
| Routing | TanStack Router (file-based) |
| UI | Mantine 9+ (PostCSS 設定済み) |
| Server state | TanStack Query |
| UI state | Zustand |
| Auth | AWS Amplify v6 (`aws-amplify` umbrella、import は `aws-amplify` / `aws-amplify/auth` / `aws-amplify/utils` で tree-shake、CLI/UI 系は使わない) |
| API クライアント | openapi-typescript + openapi-fetch |
| Search validation | valibot + `@tanstack/valibot-adapter` |
| Form | `@mantine/form` (Mantine 9 built-in `schemaResolver` で valibot を直接使う、3rd party resolver は使わない) |
| Icons | `@tabler/icons-react` |

## URL 設計

```
/sign-in                                     Cognito Managed Login UI へリダイレクト
/auth/callback                               Cognito からの戻り
/invite          ?code=AB7XQ9                招待コード入力 / 受諾
/orgs                                        所属 org 一覧 (org 選択)
/orgs/new                                    新規 org 作成
/orgs/:orgId                                 → /orgs/:orgId/transactions に redirect
/orgs/:orgId/transactions?month=YYYY-MM      取引一覧 (メイン画面)
/orgs/:orgId/transactions/new                取引追加 (Drawer)
/orgs/:orgId/transactions/:txId/edit         取引編集 (Drawer)
/orgs/:orgId/categories                      カテゴリ一覧
/orgs/:orgId/categories/new                  カテゴリ追加
/orgs/:orgId/categories/:categoryId/edit     カテゴリ編集
/orgs/:orgId/summary                         月次集計
/orgs/:orgId/members                         メンバー管理 (owner のみ)
/orgs/:orgId/settings                        org 設定 + 招待コード生成 (owner のみ)
```

TanStack Router の pathless layouts (`_public`、`_authed`) で認証ガードを分離する。

## レイアウト

スマホ前提なので AppShell の構成は次の通り。

ヘッダには org 名と切替メニューを置く。詳細画面では左に戻る矢印を出す (iOS PWA でブラウザの戻るが使えないケース対策)。フッタは下タブにして、取引・集計・カテゴリ・設定の4つを並べる。

軽い操作 (取引やカテゴリの追加・編集) は Drawer (下スライド) で開いて、× で閉じる。設定系 (メンバー管理、org 設定) は遷移を伴うフルページにし、ヘッダ左に戻る矢印を置く。

## フォーム

`@mantine/form` の `useForm` + `schemaResolver(schema, { sync: true })` を canonical pattern として採用する。schema は valibot で書く (Standard Schema 対応のため Mantine 9 built-in resolver で直接読める、`mantine-form-valibot-resolver` などの 3rd party は不要)。

```tsx
const form = useForm({
  initialValues: { name: "" },
  validate: schemaResolver(NewOrgSchema, { sync: true }),
});

<TextInput {...form.getInputProps("name")} label="組織名" required />
<form onSubmit={form.onSubmit((values) => mutation.mutate(values))}>
```

`getInputProps` が `value` / `onChange` / `error` をまとめて返すため Mantine input にそのまま spread できる。

ColorSwatch grid のような `getInputProps` に乗らない custom UI は `form.values.<field>` / `form.setFieldValue('<field>', v)` で直接読み書きする。バリデーションエラーは `form.errors.<field>` で取れる。

## カテゴリ色パレット

カテゴリ作成時の色は固定パレット (12 色) からのみ選ばせる。ユーザの自由 hex 入力は受け付けない。理由はモバイル UI で迷わない選択肢量、Mantine theme との視覚整合、表示側の統一感維持。

採用色は Mantine の各 color の shade 6。rainbow 順 (warm → cool):

<!-- prettier-ignore -->
| hex | Mantine token |
|---|---|
| `#fa5252` | red |
| `#fd7e14` | orange |
| `#fab005` | yellow |
| `#82c91e` | lime |
| `#40c057` | green |
| `#12b886` | teal |
| `#15aabf` | cyan |
| `#228be6` | blue |
| `#4c6ef5` | indigo |
| `#7950f2` | violet |
| `#be4bdb` | grape |
| `#e64980` | pink |

実体は `web-client/src/features/category/palette.ts` を SOT とする。バックエンドは hex format のみ検証する loose な validation で、過去データはパレット改訂後も表示できる。パレットへ色を増減する際はこの表と `palette.ts` を同時更新する。

色は org 内で一意 (kind を跨いで重複不可)。色 = カテゴリの視覚 ID という建付けで、`赤の食費` と `赤の給与` が並ぶと取引一覧 / chart で判別不能になる事を避ける。kind は色とは独立した軸として Badge / 符号で表現する。新規作成 UI は使用中の色を **dimmed + disabled** として表示する (非表示にしない、減ったように見えて混乱するため)。全色使用時は作成不可とする (上限 12 カテゴリ)。バックエンドは引き続き hex 形式のみ検証する (race の整合維持はフロント責務)。

## 状態管理

サーバーから取得するデータは TanStack Query で扱い、各 feature の `queries.ts` に hooks をまとめる。UI 状態 (モーダル open など) は Zustand を `stores/` に集約する。選択中の org のような URL に表れるものは TanStack Router の `params` / `search` から取る。

## OpenAPI 駆動

OpenAPI YAML はリポジトリルート直下の `openapi/openapi.yaml` を SOT として手書きする。初期は単一ファイルで進め、500行を超えたあたりから Redocly CLI で split + bundle に切り替える。

型は `web-client/src/api/schema.d.ts` に `openapi-typescript` で自動生成する。API クライアントは `openapi-fetch` で組み、パス・クエリ・レスポンスがすべて型推論される。

Web Server (Go) との同期は手動で行う (`go-playground/validator/v10` の struct tag を openapi.yaml の制約と合わせる)。本格化したら `oapi-codegen` の導入を検討する。

## フォルダ構造

`shared/` は使わずフラットに配置する。型は `import type { components } from '@/api/schema'` で参照するため、各 feature に `types.ts` は持たせない。

```
web-client/src/
├── api/
│   ├── schema.d.ts          自動生成
│   └── client.ts            openapi-fetch + Authorization 自動付与
├── components/              共通 UI (AppShell, BottomTabs, BackButton 等)
├── hooks/                   共通フック
├── stores/                  Zustand
├── theme.ts                 Mantine theme
├── routes/                  TanStack Router file-based
│   ├── __root.tsx
│   ├── _public.tsx
│   ├── _public/
│   ├── _authed.tsx
│   └── _authed/
│       └── orgs/
│           └── $orgId/
│               └── _layout/
├── features/                機能別コロケーション
│   ├── auth/                hooks.ts (Amplify ラッパー), components/
│   ├── organization/        queries.ts, components/
│   ├── transaction/         queries.ts, components/
│   ├── category/
│   ├── invite/
│   └── summary/
├── main.tsx                 Amplify.configure + Providers
└── styles.css
```

## Cognito 連携 (Amplify Auth v6)

`main.tsx` で1度だけ `Amplify.configure` を呼ぶ。

```ts
Amplify.configure({
  Auth: {
    Cognito: {
      userPoolId: import.meta.env.VITE_COGNITO_USER_POOL_ID,
      userPoolClientId: import.meta.env.VITE_COGNITO_CLIENT_ID,
      loginWith: {
        oauth: {
          domain: import.meta.env.VITE_COGNITO_DOMAIN,
          scopes: ["openid", "email"],
          redirectSignIn: [`${window.location.origin}/auth/callback`],
          redirectSignOut: [`${window.location.origin}/`],
          responseType: "code",
        },
      },
    },
  },
});
```

ログインは `signInWithRedirect()` で Managed Login UI に飛ばす。Access Token は `fetchAuthSession()` で取って openapi-fetch の Authorization ヘッダに自動付与する。

Amplify CLI / Studio / UI コンポーネントは使わない。`@aws-amplify/auth` だけ install する。
