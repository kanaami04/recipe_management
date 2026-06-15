#!/usr/bin/env bash
#
# 実機テスト用ワンコマンド起動スクリプト。
#   Postgres(docker) + バックエンド(Go, :8000) + フロント(Vite dev, :5273 を LAN 公開)
# をまとめて立ち上げ、スマホからアクセスできる URL を表示する。
#
#   使い方:  ./dev.sh
#   停止:    Ctrl+C  (バックエンド/フロントを停止。Postgres は起動したまま)
#
# 注意: このファイルは現在 Git 管理外。チームで共有したくなったら `git add dev.sh` すること。

set -euo pipefail

# このスクリプトが置かれている場所(= リポジトリのルート)
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# mise の shims を PATH に追加し、node / go / npm を mise.toml のバージョンで解決する。
export PATH="$HOME/.local/share/mise/shims:$PATH"

# air(バックエンドのホットリロード)のバイナリを解決する。
# mise の shim はグローバル版が未設定だと動かないため、インストール実体を直接探す。
AIR_BIN="$(command -v air 2>/dev/null || true)"
if [ -z "$AIR_BIN" ] || ! "$AIR_BIN" -v >/dev/null 2>&1; then
  AIR_BIN="$(ls -d "$HOME"/.local/share/mise/installs/air/*/air 2>/dev/null | sort -V | tail -1 || true)"
fi
if [ -z "$AIR_BIN" ]; then
  echo "✗ air が見つかりません。'mise use -g air@latest' でインストールしてください。" >&2
  exit 1
fi

# 指定ポートを LISTEN している node/go 系プロセスだけ停止する(VS Code 等は除外)。
# 旧 dev サーバが残ったまま再実行しても "address already in use" にならないようにする。
free_port() {
  local port="$1" pids pid
  # ポートが空(lsof が exit 1)でも set -e で落ちないよう、必ず成功させて捕捉する。
  pids="$(lsof -nP -iTCP:"$port" -sTCP:LISTEN 2>/dev/null \
    | awk 'NR>1 && ($1 ~ /node|go|main|__debug/) {print $2}' | sort -u || true)"
  for pid in $pids; do
    echo "  既存プロセス停止 pid=$pid (port $port)"
    kill "$pid" 2>/dev/null || true
  done
}

# Ctrl+C / 終了時にプロセスグループごと後始末する(npm・go・go の子バイナリをまとめて停止)。
trap 'echo; echo "▶ 停止します..."; kill 0 2>/dev/null || true' EXIT

echo "▶ 旧 dev サーバを片付け中..."
free_port 5273
free_port 8000
sleep 1

# 1) Postgres を起動(起動済みなら何もしない)
echo "▶ Postgres(docker compose)を起動..."
( cd "$ROOT/api" && docker compose up -d )

# 2) バックエンド(:8000、全インターフェース公開。air でホットリロード)
echo "▶ バックエンド(:8000, air ホットリロード)を起動..."
( cd "$ROOT/api" && "$AIR_BIN" ) &

# 3) フロント(Vite dev、--host で LAN 公開)
echo "▶ フロント(:5273, --host)を起動..."
( cd "$ROOT/frontend" && npm run dev -- --host 0.0.0.0 ) &

# LAN IP(Wi-Fi)を取得して案内を表示
IP="$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null || echo 127.0.0.1)"
sleep 2
echo
echo "================================================================"
echo "  スマホ(Mac と同じ Wi-Fi)から次の URL を開く:"
echo
echo "      http://$IP:5273"
echo
echo "  停止: このターミナルで Ctrl+C"
echo "================================================================"
echo

# どちらかのサーバが終了するまで待ち続ける(その間ログが流れる)
wait
