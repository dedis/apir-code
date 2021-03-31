// Package main defines a cli to generate the sks chunks and the database
package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/si-co/vpir-code/lib/constants"
	"github.com/si-co/vpir-code/lib/database"
	"github.com/si-co/vpir-code/lib/pgp"
	"golang.org/x/xerrors"
)

const hundredMb = 104857600
const usage = `go run main.go -cmd genChunks|genDB -path PATH -out PATH`

func main() {
	var cmd string
	var path string
	var out string

	flag.StringVar(&cmd, "cmd", "", "genChunks|genDB")
	flag.StringVar(&path, "path", "", "input file")
	flag.StringVar(&out, "out", "", "output file/folder")

	flag.Parse()

	fmt.Println(cmd, path, out)

	if cmd == "" || path == "" || out == "" {
		log.Fatalf("Usage:\n%s", usage)
	}

	switch cmd {
	case "genChunks":
		err := splitFullDumpIntoChunks(path, out)
		if err != nil {
			log.Fatalf("failed to split chunks: %v", err)
		}
	case "genDB":
		err := generateDB(path, out)
		if err != nil {
			log.Fatalf("failed to generate DB: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s", cmd)
	}
}

// func parseSksDump() {
// 	var err error
// 	fileList, err := pgp.GetSksOriginalDumpFiles(pgp.SksOriginalFolder)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	m, err := pgp.AnalyzeKeyDump(fileList)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	err = pgp.WriteKeysOnDisk(pgp.SksParsedFolder, m)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

func splitFullDumpIntoChunks(path, out string) error {
	f, err := os.Open(path)
	if err != nil {
		return xerrors.Errorf("failed to open path: %v", err)
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

			out, err := os.OpenFile(filepath.Join(out, outputName),
				os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
			if err != nil {
				return xerrors.Errorf("failed to create chunk: %v", err)
			}

			log.Printf("Writing into %s\n", outputName)

			encoder = gob.NewEncoder(out)
			numWrittenBytes = 0
			outputNum++
		}

		key := new(pgp.Key)

		// Decoding the serialized data
		if err = decoder.Decode(key); err != nil {
			if err == io.EOF {
				break
			}

			return xerrors.Errorf("failed to decode key: %v", err)
		}

		err = encoder.Encode(key)
		if err != nil {
			return xerrors.Errorf("failed to encode key: %v", err)
		}

		numWrittenBytes += len(key.Packet) + len(key.ID)
	}

	err = f.Close()
	if err != nil {
		return xerrors.Errorf("failed to close file: %v", err)
	}

	return nil
}

func generateDB(root, out string) error {
	filesInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return xerrors.Errorf("failed to read files: %v", err)
	}

	var files []string

	for _, info := range filesInfo {
		if info.IsDir() {
			continue
		}

		files = append(files, filepath.Join(root, info.Name()))
	}

	log.Printf("loading the following files: %v", files)

	db, err := database.GenerateRealKeyDB(files, constants.ChunkBytesLength, false)
	if err != nil {
		return xerrors.Errorf("failed to generate DB: %v", err)
	}

	err = db.SaveDBFileSingle(out)
	if err != nil {
		return xerrors.Errorf("failed to save db: %v", err)
	}

	return nil
}
