#!/usr/bin/env bash
# PostToolUse(Edit|Write|MultiEdit) 用。編集したファイル 1 つだけを整形する。
#
# 目的: gofmt / prettier の差分を発生源で消し、CI/レビューの整形ゲートを常に満たす。
#   リポジトリ全体の整形は遅いため、ここでは編集ファイル単体に絞る。
# 方針: 決定的な整形のみ。常に exit 0 でツールをブロックしない。
set -euo pipefail

root="${CLAUDE_PROJECT_DIR:-$(pwd)}"

file="$(jq -r '.tool_input.file_path // empty')"
[ -n "$file" ] || exit 0
[ -f "$file" ] || exit 0

# 相対パス(cwd=プロジェクト root 基準)でも解決できるよう絶対パスへ正規化する。
case "$file" in /*) abs="$file" ;; *) abs="$root/$file" ;; esac

case "$abs" in
  *.go)
    # gofmt 失敗でフックをブロックしないよう非ゼロ終了を握りつぶす。
    command -v gofmt >/dev/null 2>&1 && gofmt -w "$abs" || true
    ;;
  "$root"/frontend/*)
    case "$abs" in
      *.ts|*.tsx|*.js|*.jsx|*.json|*.css|*.html|*.md)
        # prettier は frontend の devDependency。.prettierignore は prettier 側が尊重する。
        bin="$root/frontend/node_modules/.bin/prettier"
        [ -x "$bin" ] && (cd "$root/frontend" && "$bin" --write --log-level=warn "$abs" >/dev/null 2>&1 || true)
        ;;
    esac
    ;;
esac
exit 0
