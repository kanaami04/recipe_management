import { registerSW } from 'virtual:pwa-register'

// Service Worker の登録。React Router v7 framework mode では
// vite-plugin-pwa の自動注入が効かないため手動で登録する(ワークアラウンド②)。
// autoUpdate 戦略のため onNeedRefresh のハンドリングは不要。
registerSW({ immediate: true })
