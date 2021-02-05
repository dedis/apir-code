package main

import (
	"log"

	"github.com/nikirill/go-crypto/openpgp"
	"github.com/si-co/vpir-code/lib/pgp"
)

func main() {
	var m map[string]*openpgp.Entity
	var err error
	fileList, err := pgp.GetSksDumpFiles(pgp.SksOriginalFolder)
	if err != nil {
		log.Fatal(err)
	}
	m, err = pgp.AnalyzeDumpFiles(fileList)
	if err != nil {
		log.Fatal(err)
	}
	err = pgp.WriteKeysOnDisk(pgp.SksDestinationFolder, m)
	if err != nil {
		log.Fatal(err)
	}
}
