import type { Config } from '@react-router/dev/config'

// React Router framework mode を SPA 構成で使う。
// ssr:false で静的配信のまま loader/action・型安全ルートの恩恵を受ける。
// appDirectory は既存コードに合わせて src のままにする。
export default {
  appDirectory: 'src',
  ssr: false,
  // React Router v8 の新挙動へ先行 opt-in する(dev の Future Flag 警告を解消)。
  future: {
    v8_middleware: true,
    v8_splitRouteModules: true,
    v8_viteEnvironmentApi: true,
    v8_passThroughRequests: true,
    v8_trailingSlashAwareDataRequests: true,
  },
} satisfies Config
