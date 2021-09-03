package utils

// adapted from https://github.com/henrycg/prio/master/utils/profile.go

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
)

func StartProfiling(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)

	// Stop on ^C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	go func() {
		for range c {
			// sig is a ^C, handle it
			pprof.StopCPUProfile()
			os.Exit(0)
		}
	}()
}

func StopProfiling() {
	// Stop when process exits
	pprof.StopCPUProfile()
}

func writeMemProfile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Writing memory profile")
	pprof.WriteHeapProfile(f)
	f.Close()
}

func StartMemProfiling(filename string) {
	// Stop on ^C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			writeMemProfile(filename)
			os.Exit(0)
		}
	}()
}

func writeBlockProfile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Writing block profile")
	pprof.Lookup("block").WriteTo(f, 0)
	f.Close()
}

func StartBlockProfiling(filename string) {
	// Stop on ^C
	runtime.SetBlockProfileRate(1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
			writeBlockProfile(filename)
			os.Exit(0)
		}
	}()
}
