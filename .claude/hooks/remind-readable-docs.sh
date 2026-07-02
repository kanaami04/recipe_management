#!/usr/bin/env bash
# PostToolUse(Edit|Write|MultiEdit) 用。Markdown ドキュメントを編集したとき、
# readable-docs スキルの作法に沿っているか確認するよう Claude に促す。
#
# 目的: readable-docs のスキル自動起動は確率的で、素の .md 編集では発火しにくい。
#   フックで決定的にリマインドし、文章作法の適用漏れを下支えする。
# 方針: ブロックしない。additionalContext で助言するだけ(常に exit 0)。
#   ノイズを避けるため、1 セッションにつき最初の 1 回だけ通知する。
set -euo pipefail

# stdin は 1 度しか読めないため、file_path と session_id を 1 回で取り出す。
input="$(cat)"
file="$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')"
session="$(printf '%s' "$input" | jq -r '.session_id // "nosession"')"
[ -n "$file" ] || exit 0

# .md のみ対象。生成物・依存ディレクトリは除外。
case "$file" in
  *.md) ;;
  *) exit 0 ;;
esac
case "$file" in
  node_modules/*|*/node_modules/*) exit 0 ;;
esac

# 1 セッション 1 回に絞る(読み込めばコンテキストに残るため毎回は不要)。
marker="${TMPDIR:-/tmp}/readable-docs-reminded.${session}"
[ -e "$marker" ] && exit 0
: > "$marker" 2>/dev/null || true

jq -n '{
  hookSpecificOutput: {
    hookEventName: "PostToolUse",
    additionalContext: "Markdown ドキュメントを編集しました。readable-docs スキルをまだ読んでいなければ Skill ツールで読み込み、その文章作法に沿っているか確認して必要なら整えてください。"
  }
}'
exit 0
