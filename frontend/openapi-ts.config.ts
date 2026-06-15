import { defineConfig } from '@hey-api/openapi-ts'

// api/openapi.yaml を単一ソースに、TS 型・zod・TanStack Query フックを生成する
// (frontend ADR-0007 / api ADR-0005)。生成物は src/shared/api/generated/ に置く。
export default defineConfig({
  input: '../api/openapi.yaml',
  output: {
    path: './src/shared/api/generated',
  },
  plugins: [
    '@hey-api/client-axios',
    '@hey-api/typescript',
    '@hey-api/sdk',
    // API レスポンスを実行時に検証する zod を生成(コンパイル時 + 実行時の二重防御, ADR-0007 #3)
    {
      name: 'zod',
      responses: true,
    },
    // TanStack Query の queryOptions / mutationOptions を生成(ADR-0003 / ADR-0007 #2)
    {
      name: '@tanstack/react-query',
    },
  ],
})
