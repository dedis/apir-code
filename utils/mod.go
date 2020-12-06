package utils

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"golang.org/x/xerrors"
	"strconv"
)

type serverConfig struct {
	Servers map[string]server
}

type server struct {
	Ip string
	Port int
}

func LoadServerConfig(configFile string) ([]string, error) {
	sc := serverConfig{
		Servers: make(map[string]server),
	}
	_, err := toml.DecodeFile(configFile, &sc)
	if err != nil {
		return nil, xerrors.Errorf("toml decoding: %v", err)
	}
	addresses := make([]string, len(sc.Servers))
	for index, server := range sc.Servers {
		i, err := strconv.Atoi(index)
		if err != nil {
			return nil, xerrors.Errorf("could not convert server index to integer: %v", err)
		}
		addresses[i] = fmt.Sprintf("%s:%d", server.Ip, server.Port)
	}
	return addresses, nil
}
