#!/usr/bin/env bash
# PreToolUse(Edit|Write|MultiEdit) 用。生成コードの手編集をブロックする。
#
# 対象:
#   - api/internal/apigen/apigen.gen.go   (oapi-codegen の出力)
#   - frontend/src/shared/api/generated/** (openapi-ts の出力)
# いずれも openapi.yaml を単一ソースに生成する。手で直しても次の再生成で消える。
#
# 拒否は exit 0 + JSON(permissionDecision=deny) が正規(exit 2 だと JSON が解析されない)。
set -euo pipefail

file="$(jq -r '.tool_input.file_path // empty')"

case "$file" in
  */api/internal/apigen/apigen.gen.go|api/internal/apigen/apigen.gen.go| \
  */frontend/src/shared/api/generated/*|frontend/src/shared/api/generated/*)
    cat <<'JSON'
{
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "deny",
    "permissionDecisionReason": "生成コード (api/internal/apigen, frontend/src/shared/api/generated) は手編集しないでください。openapi.yaml を変更し、api は `mise run gen-api`、frontend は `npm run gen:api` で再生成してください。"
  }
}
JSON
    ;;
esac
exit 0
