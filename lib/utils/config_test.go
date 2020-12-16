package utils

import (
	"fmt"
	"testing"
)

func TestLoad(t *testing.T) {
	fmt.Println(LoadConfig("../../config.toml"))
}
