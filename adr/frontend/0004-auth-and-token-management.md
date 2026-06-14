# 0004. 認証とトークン管理: メモリ保持の access + httpOnly Cookie の refresh

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

現状はログイン時に取得した access / refresh を **localStorage** に保存し、
fetcher や各コンポーネントで手動に `Authorization` ヘッダを付けている。トークン更新(refresh)も
401 ハンドリングも無く、トークンが切れると無言で壊れる。localStorage は JS から読めるため、
XSS が起きると **7 日有効の refresh トークンを盗み出される**(被害が長期化する)。

バックエンドは Bearer トークン方式(`POST /api/token/` が JSON で access/refresh を返し、
`POST /api/token/refresh/` が refresh から新 access を返す。TTL は access 1h / refresh 7d)。

SPA で「正しく」やるなら、長期トークンを JS から触れない場所に置くのがベストプラクティス。

## 決定

### 1. access はメモリ、refresh は httpOnly Cookie

- **access**: JSON body で受け取り、**メモリ保持**(永続化しない)。
- **refresh**: バックエンドが `HttpOnly; Secure; SameSite` Cookie で発行する。
  JS から読めないため、XSS でも長期トークンを持ち出せない(被害がセッション中に限定される)。

これはバックエンドの改修を伴うため、対になる [api ADR-0004](../api/0004-cookie-based-refresh-token.md) を併せて定める。

### 2. 同一オリジンに寄せる(Cookie 方式の前提)

フロントと API が別オリジンだと Cookie 送信に `SameSite=None` が必要になり CSRF リスクが上がる。
これを避けるため同一オリジンに寄せる。

- **開発**: Vite の dev proxy で `/api` を Go バックエンドへ転送し、同一オリジン化する
  (CORS 不要・`SameSite=Lax` で Cookie が動く)。
- **本番**: リバースプロキシ等でフロントと API を同一サイトに配置する。

baseURL のハードコードは廃止し、`/api` 相対で叩く([ADR-0009](0009-coding-conventions.md))。

### 3. axios interceptor に集約する

- **リクエスト interceptor**: access を自動付与する(手動ヘッダ付けを廃止)。
- **レスポンス interceptor**: 401 を受けたら refresh で更新 → 元リクエストを再試行 →
  refresh も失敗ならログアウトする。
- **single-flight**: 同時多発の 401 で refresh が複数回走らないよう 1 回に集約する。

### 4. access の置き場所は loader からも読めるストアにする

`clientLoader` は React の外側で動くため([ADR-0002](0002-routing-react-router-framework-mode.md))、
access を Context だけに置くと loader から読めない。**モジュールレベルの小さなストア**
(プレーン変数 or 軽量ストア)に保持し、axios interceptor と clientLoader の双方がそこを参照する。
表示用に Context へミラーする。

### 5. 認証ガードは clientLoader、CSRF は同一オリジン + SameSite で守る

- 保護ルートの `clientLoader` で未ログインを `redirect("/")`([ADR-0002](0002-routing-react-router-framework-mode.md))。
- CSRF は「同一オリジン + `SameSite=Lax` + カスタムヘッダ必須」で実用上十分とする。
  将来クロスサイト構成が必要になったら double-submit token を後付けする。

## 結果

### 良い点

- 長期 refresh トークンが JS から触れない場所に移り、XSS の被害が大幅に縮小する。
- トークン付与・更新・401 が interceptor に集約され、コンポーネントから消える。
- 認証ガードが loader に集約され、現状のグローバル可変変数バグが解消する。

### トレードオフ

- バックエンド改修(Cookie 発行・CORS・CSRF)が必要([api ADR-0004](../api/0004-cookie-based-refresh-token.md))。
- 同一オリジン化のため dev proxy / 本番プロキシの構成管理が増える。
- リロード直後は access が無く、refresh で再取得する一手間が要る。
