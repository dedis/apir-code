package main

import (
	"log"

	"github.com/nikirill/go-crypto/openpgp"
	"github.com/si-co/vpir-code/lib/pgp"
)

func main() {
	var m map[string]*openpgp.Entity
	var err error
	fileList := []string{"sks-dump/sks-dump-0000.pgp"}
	m, err = pgp.AnalyzeDumpFiles(fileList)
	if err != nil {
		log.Fatal(err)
	}
	err = pgp.WriteKeysOnDisk("sks/", m)
	if err != nil {
		log.Fatal(err)
	}
}
