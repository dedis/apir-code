package utils

import (
	"fmt"
	"log"
	"strconv"

	"github.com/BurntSushi/toml"
	"golang.org/x/xerrors"
)

type Config struct {
	Servers map[string]Server
	Creds   Creds

	Addresses []string

	ClientCertFile string
	ClientKeyFile  string
}

type Server struct {
	Index int
	IP    string
	Port  int
}

type Creds struct {
	CertificateFile string
	KeyFile         string
}

func LoadConfig(configFile string) (*Config, error) {
	log.Printf("Loading config file from %s", configFile)
	if configFile == "" {
		log.Fatalf("config file is not set")
	}

	// load config file
	c := new(Config)
	_, err := toml.DecodeFile(configFile, c)
	if err != nil {
		return nil, xerrors.Errorf("toml decoding: %v", err)
	}

	// parse and store server addresses
	addresses := make([]string, len(c.Servers))
	for index, server := range c.Servers {
		i, err := strconv.Atoi(index)
		if err != nil {
			return nil, xerrors.Errorf("could not convert server index to integer: %v", err)
		}
		addresses[i] = fmt.Sprintf("%s:%d", server.IP, server.Port)
	}
	c.Addresses = addresses

	return c, nil
}
