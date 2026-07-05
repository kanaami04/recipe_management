// Package invite はシェアグループの招待コードを生成する。
package invite

import (
	"crypto/rand"
	"math/big"
)

// alphabet は紛らわしい文字(0/O, 1/I など)を除いた英数字。
const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// codeLength は招待コードの文字数。
const codeLength = 8

// Code は暗号乱数から 8 文字の招待コードを生成する。
func Code() (string, error) {
	b := make([]byte, codeLength)
	max := big.NewInt(int64(len(alphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}
