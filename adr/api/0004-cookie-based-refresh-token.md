# 0004. リフレッシュトークンの Cookie 化と CORS/CSRF 方針

- ステータス: Accepted
- 日付: 2026-06-14

## コンテキスト

現状、`POST /api/token/` は access / refresh を JSON body で返し、フロントが localStorage に保存している。
refresh は 7 日有効で、JS から読める場所にあるため XSS で持ち出されると被害が長期化する。

フロントは「access をメモリ保持、refresh を httpOnly Cookie に」というベストプラクティスへ移行する
([frontend ADR-0004](../frontend/0004-auth-and-token-management.md))。これを実現するため、
バックエンド側の Cookie 発行・CORS・CSRF の方針を定める。

## 決定

### 1. refresh は httpOnly Secure Cookie で発行する

- `POST /api/token/` は **access を JSON body** で返し、**refresh を `HttpOnly; Secure; SameSite=Lax` Cookie**
  でセットする(body には refresh を含めない方向)。
- `POST /api/token/refresh/` は **Cookie から refresh を読み**、新しい access を JSON body で返す
  (現状の「body の `{refresh}`」依存から Cookie へ移す)。
- Cookie の `Path` は refresh エンドポイントに絞る。`Max-Age` は refresh TTL(7 日)に合わせる。
- ログアウト用に Cookie を失効させるエンドポイントを設ける。

### 2. 同一オリジン運用を前提にする

フロントと API を同一サイトに寄せ(dev は Vite proxy、本番はリバースプロキシ。
[frontend ADR-0004](../frontend/0004-auth-and-token-management.md))、`SameSite=Lax` で運用する。
クロスサイト(`SameSite=None`)は CSRF リスクのため避ける。CORS が必要な構成では
`Access-Control-Allow-Credentials: true` と特定オリジン許可(ワイルドカード禁止)を行う。

### 3. CSRF 対策

Cookie 方式になるため CSRF 対策を加える。「同一オリジン + `SameSite=Lax` + カスタムヘッダ必須」を基本とし、
将来クロスサイトが必要になったら double-submit token を導入する。

### 4. エラーマッピングは既存方針を踏襲する

refresh の失効・改竄は 401 を返す([ADR-0003](0003-error-handling.md) のマッピングに従う)。

## 結果

### 良い点

- 長期 refresh トークンが JS から触れなくなり、XSS 被害を縮小できる。
- access は短命・メモリ保持となり、漏洩時の影響範囲が狭まる。

### トレードオフ

- Cookie 発行・読み取り、CORS credentials、CSRF 対策の実装が増える。
- 同一オリジン運用のためのプロキシ構成管理が必要になる。
- access を body・refresh を Cookie と扱いが分かれ、認証フローがやや複雑になる。
