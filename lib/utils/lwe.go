package utils

import (
	"compress/gzip"
	"encoding/gob"
	"os"
)

func WriteMatrixToFile(filename string, in []uint32) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fz := gzip.NewWriter(f)
	defer fz.Close()

	encoder := gob.NewEncoder(fz)
	err = encoder.Encode(in)
	if err != nil {
		return err
	}

	return nil
}

func LoadMatrixFromFile(filename string) ([]uint32, error) {
	out := make([]int, 0)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer fz.Close()

	decoder := gob.NewDecoder(fz)
	err = decoder.Decode(&out)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
