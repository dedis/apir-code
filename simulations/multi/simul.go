package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yahoo/vssh"
)

func main() {
	vs := vssh.New().Start()
	config := vssh.GetConfigUserPass(os.Getenv("APIR_USERNAME"), os.Getenv("APIR_PASSWORD"))
	for _, addr := range []string{"iccluster107.iccluster.epfl.ch:22", "iccluster106.iccluster.epfl.ch:22"} {
		vs.AddClient(addr, config, vssh.SetMaxSessions(4))
	}
	vs.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := "ping -c 4 1.1.1.1"
	timeout, _ := time.ParseDuration("6s")
	respChan := vs.Run(ctx, cmd, timeout)

	for resp := range respChan {
		if err := resp.Err(); err != nil {
			log.Println(err)
			continue
		}

		outTxt, errTxt, _ := resp.GetText(vs)
		fmt.Println(outTxt, errTxt, resp.ExitStatus())
	}

}
