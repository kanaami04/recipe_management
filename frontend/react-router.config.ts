import type { Config } from '@react-router/dev/config'

// React Router framework mode を SPA 構成で使う (frontend ADR-0002 / ADR-0001)。
// ssr:false で静的配信のまま loader/action・型安全ルートの恩恵を受ける。
// appDirectory は既存コードに合わせて src のままにする(feature 再編は ADR-0005 で別途)。
export default {
  appDirectory: 'src',
  ssr: false,
} satisfies Config
