package database

import (
	"github.com/si-co/vpir-code/lib/constants"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateRealKeysDB(t *testing.T) {
	sksFilesPath := "../../data/sks/"
	nRows := 1
	_, err := GenerateRealKeyDB(sksFilesPath, constants.ChunkBytesLength, false)
	require.NoError(t, err)
}

func TestPadBlock(t *testing.T) {
	b := []byte{0x01, 0xff, 0x35}
	b = PadBlock(b)
	require.Equal(t, []byte{0x01, 0xff, 0x35, 0x80}, b)
}

func TestUnPadBlock(t *testing.T) {
	b := []byte{0x01, 0xff, 0x35, 0x80, 0x00, 0x00, 0x00}
	b = UnPadBlock(b)
	require.Equal(t, []byte{0x01, 0xff, 0x35}, b)
}
