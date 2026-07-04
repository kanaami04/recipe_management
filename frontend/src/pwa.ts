import { registerSW } from 'virtual:pwa-register'

// 開きっぱなしのタブ向けの定期チェック間隔(1 時間)。
const UPDATE_INTERVAL_MS = 60 * 60 * 1000

// Service Worker の登録。React Router v7 framework mode では
// vite-plugin-pwa の自動注入が効かないため手動で登録する(ワークアラウンド②)。
// autoUpdate 戦略のため、新版を見つけたら自動でリロードして反映する
// (onNeedRefresh のハンドリングは不要)。
//
// 既定の更新チェックはページ遷移時やブラウザ任せ(最長 24h)で、バックグラウンドから
// 復帰しただけでは走らない。デプロイをなるべく早く反映させるため、明示的に呼ぶ:
//   - フォアグラウンド復帰(visibilitychange → visible)時
//   - 一定間隔(長時間開いたままのタブ向け)
registerSW({
  immediate: true,
  onRegisteredSW(_swScriptUrl, registration) {
    if (!registration) return
    const checkForUpdate = () => {
      // オフライン等の一時的な失敗は無視する(次の機会に再チェックされる)。
      registration.update().catch(() => undefined)
    }
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'visible') checkForUpdate()
    })
    setInterval(checkForUpdate, UPDATE_INTERVAL_MS)
  },
})
