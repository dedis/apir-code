package utils

import (
	"fmt"
	"strconv"

	"github.com/BurntSushi/toml"
	"golang.org/x/xerrors"
)

type Config struct {
	DBLength int
	Servers  map[string]Server
}

type Server struct {
	Index int
	IP    string
	Port  int
}

func LoadConfig(configFile string) (*Config, error) {
	c := new(Config)
	_, err := toml.DecodeFile(configFile, c)
	if err != nil {
		return nil, xerrors.Errorf("toml decoding: %v", err)
	}

	return c, nil
}

func ServerAddresses(c *Config) ([]string, error) {
	addresses := make([]string, len(c.Servers))
	for index, server := range c.Servers {
		i, err := strconv.Atoi(index)
		if err != nil {
			return nil, xerrors.Errorf("could not convert server index to integer: %v", err)
		}
		addresses[i] = fmt.Sprintf("%s:%d", server.IP, server.Port)
	}

	return addresses, nil
}
