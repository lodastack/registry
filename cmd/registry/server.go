package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/cluster"
	"github.com/lodastack/registry/httpd"
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

func join(joinAddr, raftAddr string) error {
	// Join using IP address, as that is what Hashicorp Raft works in.
	resv, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		return err
	}

	// Check for protocol scheme, and insert default if necessary.
	fullAddr := httpd.NormalizeAddr(fmt.Sprintf("%s/api/v1/peer", joinAddr))

	// Enable skipVerify as requested.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	for {
		b, err := json.Marshal(map[string]string{"addr": resv.String()})
		if err != nil {
			return err
		}

		// Attempt to join.
		resp, err := client.Post(fullAddr, "application-type/json", bytes.NewReader(b))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		switch resp.StatusCode {
		case http.StatusOK:
			return nil
		case http.StatusMovedPermanently:
			fullAddr = resp.Header.Get("location")
			if fullAddr == "" {
				return fmt.Errorf("failed to join, invalid redirect received")
			}
			log.Printf("join request redirecting to", fullAddr)
			continue
		default:
			return fmt.Errorf("failed to join, node returned: %s: (%s)", resp.Status, string(b))
		}
	}
}

func publishAPIAddr(c *cluster.Service, raftAddr, apiAddr string, t time.Duration) error {
	tck := time.NewTicker(publishPeerDelay)
	defer tck.Stop()
	tmr := time.NewTimer(t)
	defer tmr.Stop()

	for {
		select {
		case <-tck.C:
			if err := c.SetPeer(raftAddr, apiAddr); err != nil {
				log.Errorf("failed to set peer for %s to %s: %s (retrying)",
					raftAddr, apiAddr, err.Error())
				continue
			}
			return nil
		case <-tmr.C:
			return fmt.Errorf("set peer timeout expired")
		}
	}
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
		pprof.Lookup("heap").WriteTo(prof.mem, 0)
		prof.mem.Close()
		log.Printf("memory profiling stopped")
	}
}
