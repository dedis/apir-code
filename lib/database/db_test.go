package database

import (
	"fmt"
	"testing"
)

func TestCreateMatrix(*testing.T) {
	db := CreateAsciiMatrix()
	fmt.Println(db)
}
