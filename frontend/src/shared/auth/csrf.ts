// CSRF 対策のカスタムヘッダ。
// クロスサイトのフォーム送信はカスタムヘッダを付けられないため、
// 全リクエストに付与してバックエンドの必須検証を満たす。
export const CSRF_HEADER = 'X-Requested-With'
export const CSRF_HEADER_VALUE = 'XMLHttpRequest'
