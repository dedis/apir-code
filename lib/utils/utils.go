package utils

import (
	"golang.org/x/xerrors"
)

func BitStringToBytes(s string) ([]byte, error) {
	b := make([]byte, (len(s)+(8-1))/8)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '1' {
			return nil, xerrors.New("not a bit")
		}
		b[i>>3] |= (c - '0') << uint(7-i&7)
	}
	return b, nil
}
