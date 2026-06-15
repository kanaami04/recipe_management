// access トークンのメモリ保持ストア (frontend ADR-0004)。
//
// localStorage に置かず変数で保持する(リロードで消え、refresh Cookie から再取得する)。
// clientLoader は React の外側で動くため、Context ではなくこのモジュールを参照する。
// 表示用に Context へミラーする(subscribe)。

let accessToken: string | null = null

type Listener = (token: string | null) => void
const listeners = new Set<Listener>()

export function getAccessToken(): string | null {
  return accessToken
}

export function setAccessToken(token: string | null): void {
  accessToken = token
  listeners.forEach((listener) => listener(token))
}

export function clearAccessToken(): void {
  setAccessToken(null)
}

// 変更通知を購読する。返り値で解除する。
export function subscribeAccessToken(listener: Listener): () => void {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}
