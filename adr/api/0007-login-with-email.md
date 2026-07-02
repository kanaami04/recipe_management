# 0007. ログイン識別子を username から email に変更する

- ステータス: Accepted
- 日付: 2026-07-02

## コンテキスト

ログイン(`POST /api/token/`)は Django 時代の名残で username + password を受け取っていた。
一方、ユーザーが日常的に覚えているのは email であり、username は表示名的な役割しか持たない
(共有先の指定などにしか使っていない)。login の識別子は unique かつユーザーが確実に
記憶しているものが望ましい。email はすでに `users.email` に unique 制約があり、登録時に
必須・形式検証もしている。

## 決定

### 1. ログインの識別子を email にする

- `TokenRequest` を `{ email, password }` に変更(`username` を廃止)。email は
  `format: email` + `validate: required,email` で形式検証する。
- `AuthService.Login(ctx, email, password)` は `UserRepository.FindByEmail` で引く
  (従来は `FindByUsername`)。パスワード照合・無効ユーザー扱いは変更なし。
- username は登録(`RegisterRequest`)・共有先指定・表示には引き続き使う。ログインの
  識別子から外すだけで、フィールド自体は残す。

### 2. フロントもログインフォームを email 入力にする

- `loginFormSchema` を `{ email, password }`(email は必須 + 形式検証)に変更。
- `LoginForm` の入力を `type="email"`・ラベル「メールアドレス」に変更し、
  `authClient.login(email, password)` が `{ email, password }` を送る。

## 結果

- ログイン識別子が unique 制約付きの email に統一され、意味的にも自然になった。
- 破壊的変更だが、利用者が少なく既存トークンにも影響しないため移行措置は取らない。
- username は引き続き存在するため、共有先指定・表示などの既存機能に影響はない。
