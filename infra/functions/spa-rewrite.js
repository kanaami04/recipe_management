// SPA フォールバック用の CloudFront Function (viewer-request, adr/infra/0001 #4)。
// 拡張子を含まない URI (= SPA のルート) を /index.html に書き換えて S3 へ向ける。
// /sw.js や /manifest.webmanifest は「.」を含むため素通りし、/api/* は
// 別 behavior が先にマッチするためこの関数を通らない。
// custom error response を使わない理由: distribution 全体に効いてしまい、
// /api/* の 403/404 レスポンスまで index.html に化けるため。
function handler(event) {
  var uri = event.request.uri;
  if (!uri.includes('.')) {
    event.request.uri = '/index.html';
  }
  return event.request;
}
