package main

import (
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/model"
)

func initLog(dir string, level string, rotatenum int, size uint64) error {
	var err error
	model.LogBackend, err = log.NewFileBackend(dir)
	if err != nil {
		return err
	}
	log.SetLogging(level, model.LogBackend)
	log.Rotate(rotatenum, size)
	return nil
}

// prof stores the file locations of active profiles.
var prof struct {
	cpu *os.File
	mem *os.File
}

// startProfile initializes the CPU and memory profile, if specified.
func startProfile(cpuprofile, memprofile string) {
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Errorf("failed to create CPU profile file at %s: %s", cpuprofile, err.Error())
		}
		log.Printf("writing CPU profile to: %s\n", cpuprofile)
		prof.cpu = f
		pprof.StartCPUProfile(prof.cpu)
	}

	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Errorf("failed to create memory profile file at %s: %s", cpuprofile, err.Error())
		}
		log.Printf("writing memory profile to: %s\n", memprofile)
		prof.mem = f
		runtime.MemProfileRate = 4096
	}
}

// stopProfile closes the CPU and memory profiles if they are running.
func stopProfile() {
	if prof.cpu != nil {
		pprof.StopCPUProfile()
		prof.cpu.Close()
		log.Printf("CPU profiling stopped")
	}
	if prof.mem != nil {
		runtime.GC()
		if err := pprof.Lookup("heap").WriteTo(prof.mem, 0); err != nil {
			log.Errorf("could not write memory profile: %s\n", err)
		}
		prof.mem.Close()
		log.Printf("memory profiling stopped")
	}
}
