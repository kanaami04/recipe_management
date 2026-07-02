// Package id は全エンティティの主キーに使う ID を採番する。
package id

import "github.com/google/uuid"

// New は UUIDv7 を文字列で返す。
// UUIDv7 は先頭にミリ秒精度のタイムスタンプを含むため時刻順にソートでき、
// ランダムな v4 と違って B-Tree インデックスの断片化が起きにくい。
//
// uuid.NewV7 は乱数源(crypto/rand)の読み取りに失敗したときだけエラーを返す。
// これはプロセスが継続不能な致命的状態なので Must で panic させる。
func New() string {
	return uuid.Must(uuid.NewV7()).String()
}
