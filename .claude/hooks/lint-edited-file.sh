#!/usr/bin/env bash
# PostToolUse(Edit|Write|MultiEdit) 用。編集した frontend ファイルを eslint で助言的に検査する。
#
# 方針: ブロックしない。指摘があれば additionalContext で Claude に伝えるだけ。
#   go の静的検査(go vet)はパッケージ単位で粗く遅いため CI/手動に委ね、ここでは扱わない。
#   format-edited-file.sh の後に走る前提(settings.json の配列順で整形 → lint)。
set -euo pipefail

root="${CLAUDE_PROJECT_DIR:-$(pwd)}"

file="$(jq -r '.tool_input.file_path // empty')"
[ -n "$file" ] || exit 0
[ -f "$file" ] || exit 0

case "$file" in /*) abs="$file" ;; *) abs="$root/$file" ;; esac

# frontend 配下の TS/JS/TSX のみ対象。
case "$abs" in
  "$root"/frontend/*) ;;
  *) exit 0 ;;
esac
case "$abs" in
  *.ts|*.tsx|*.js|*.jsx) ;;
  *) exit 0 ;;
esac

bin="$root/frontend/node_modules/.bin/eslint"
[ -x "$bin" ] || exit 0

out="$(cd "$root/frontend" && "$bin" "$abs" 2>&1 || true)"
[ -n "$out" ] || exit 0

jq -n --arg ctx "eslint の指摘 ($abs):
$out" '{hookSpecificOutput: {hookEventName: "PostToolUse", additionalContext: $ctx}}'
exit 0
