package utils

// adapted from https://github.com/henrycg/prio/master/utils/profile.go

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
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
		for _ = range c {
			// sig is a ^C, handle it
			writeMemProfile("client-mem.prof")
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
	runtime.GC()
	pprof.WriteHeapProfile(f)
	f.Close()
}

func StartMemProfiling(filename string) {
	// Stop on ^C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
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
	// go func() {
	// 	for {
	// 		time.Sleep(time.Second * 3)
	// 		writeBlockProfile(filename)
	// 	}
	// }()
	// Stop on ^C
	runtime.SetBlockProfileRate(1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			// sig is a ^C, handle it
			writeBlockProfile(filename)
			// os.Exit(0)
		}
	}()
}

func StartTrace(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("failed to create trace output file: %v", err)
	}

	c := make(chan os.Signal, 1)

	go func() {
		for range c {
			// sig is a ^C, handle it
			if err := f.Close(); err != nil {
				log.Fatalf("failed to close trace file: %v", err)
			}
		}
	}()

	if err := trace.Start(f); err != nil {
		log.Fatalf("failed to start trace: %v", err)
	}
	defer trace.Stop()
}
