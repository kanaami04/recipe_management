// Package testutil はテスト用の共通ヘルパーを提供する。
package testutil

import (
	"os"
	"testing"
)

// EnvRunIntegration が "1" のときだけ結合テストを実行する。
const EnvRunIntegration = "RUN_INTEGRATION"

// RequireIntegration は RUN_INTEGRATION=1 でなければテストを Skip する。
// 実 DB（testcontainers）を使う結合テストの先頭で呼ぶ。
func RequireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv(EnvRunIntegration) != "1" {
		t.Skipf("skipping integration test; set %s=1 to run", EnvRunIntegration)
	}
}
