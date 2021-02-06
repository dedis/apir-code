package main

import (
	"encoding/gob"
	"fmt"
	"github.com/si-co/vpir-code/lib/pgp"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	sksDir    = "data/sks"
	hundredMb = 104857600
)

func main() {
	splitFullDumpIntoChunks(pgp.SksParsedFullFileName)
	fmt.Println("DONE")
}

func parseSksDump() {
	var err error
	fileList, err := pgp.GetSksOriginalDumpFiles(pgp.SksOriginalFolder)
	if err != nil {
		log.Fatal(err)
	}
	m, err := pgp.AnalyzeKeyDump(fileList)
	if err != nil {
		log.Fatal(err)
	}
	err = pgp.WriteKeysOnDisk(pgp.SksDestinationFolder, m)
	if err != nil {
		log.Fatal(err)
	}
}

func splitFullDumpIntoChunks(file string) {
	var err error
	f, err := os.Open(filepath.Join(sksDir, file))
	if err != nil {
		log.Fatal(err)
	}
	decoder := gob.NewDecoder(f)

	var encoder *gob.Encoder
	numWrittenBytes := 0
	outputNum := 0
	var outputName string
	for {
		if encoder == nil || numWrittenBytes > hundredMb {
			// If the file already exists, the content is overwritten
			outputName = fmt.Sprintf("sks-%03d.pgp", outputNum)
			out, err := os.OpenFile(filepath.Join(sksDir, outputName),
				os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Writing into %s\n", outputName)
			encoder = gob.NewEncoder(out)
			numWrittenBytes = 0
			outputNum += 1
		}
		key := new(pgp.Key)
		// Decoding the serialized data
		if err = decoder.Decode(key); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		if err = encoder.Encode(key); err != nil {
			log.Fatal(err)
		}
		numWrittenBytes += len(key.Packet) + len(key.ID)
	}
	if err = f.Close(); err != nil {
		log.Fatal(err)
	}
}
