import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import simpleImportSort from 'eslint-plugin-simple-import-sort'
import tseslint from 'typescript-eslint'
import { globalIgnores } from 'eslint/config'

export default tseslint.config([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs['recommended-latest'],
      reactRefresh.configs.vite,
    ],
    plugins: {
      'simple-import-sort': simpleImportSort,
    },
    languageOptions: {
      // tsconfig (ES2022) と揃える (ADR-0009 #7)
      ecmaVersion: 2022,
      globals: globals.browser,
    },
    rules: {
      // import / export 順を機械的に統一する (ADR-0009 #7)
      'simple-import-sort/imports': 'error',
      'simple-import-sort/exports': 'error',
      // console.log は禁止。warn/error と専用 logger のみ許容 (ADR-0009 #2)
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      // alert/confirm/prompt は廃止しトースト・ダイアログへ (ADR-0009 #3)
      'no-alert': 'warn',
      // 定数 export (cva など) は fast-refresh の対象外として許可する
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    },
  },
  {
    // テストヘルパー・テストは dev の Fast Refresh グラフに乗らないため、
    // 非コンポーネントの export(RTL 再エクスポート等)を許可する。
    files: ['src/test/**', '**/*.test.{ts,tsx}'],
    rules: {
      'react-refresh/only-export-components': 'off',
    },
  },
])
