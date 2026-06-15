import { redirect } from 'react-router'

// 未定義パスはログインへリダイレクトする。
// clientLoader は React の外側で動くため、ここで redirect を返す (ADR-0002)。
export function clientLoader() {
  return redirect('/')
}

export default function CatchAll() {
  return null
}
