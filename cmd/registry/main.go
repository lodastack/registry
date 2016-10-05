package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/lodastack/registry/cluster"
	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/httpd"
	"github.com/lodastack/registry/store"
	"github.com/lodastack/registry/tcp"
)

// Command line defaults
const (
	DefaultConfigFile = "/etc/registry/registry.conf"

	publishPeerDelay   = 1 * time.Second
	publishPeerTimeout = 30 * time.Second
)

// Command line parameters
var configFile string
var joinAddr string

const (
	muxRaftHeader = 1 // Raft consensus communications
	muxMetaHeader = 2 // Cluster meta communications
)

func init() {
	flag.StringVar(&configFile, "config", DefaultConfigFile, "Set the config file")
	flag.StringVar(&joinAddr, "join", "", "Set the leader API addr to join a cluster")
}

func main() {
	flag.Parse()

	//parse config file
	err := config.ParseConfig(configFile)
	if err != nil {
		fmt.Println("Parse Config File Error : " + err.Error())
		os.Exit(1)
	}

	//save pid to file
	err = ioutil.WriteFile(config.C.CommonConf.PID, []byte(strconv.Itoa(os.Getpid())), 0744)
	if err != nil {
		fmt.Println("write PID file error: ", err)
		os.Exit(1)
	}

	// store config
	c := config.C.DataConf
	// TODO: remove joinaddr from config file
	if joinAddr == "" {
		joinAddr = c.ClusterLeader
	}

	// serve mux TCP
	ln, err := net.Listen("tcp", c.ClusterBind)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", c.ClusterBind, err.Error())
		os.Exit(1)
	}
	mux := tcp.NewMux(ln, nil)
	go mux.Serve()

	// Start up mux and get transports for cluster.
	raftTn := mux.Listen(muxRaftHeader)
	s := store.New(c.Dir, raftTn)
	if err := s.Open(joinAddr == ""); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	// Create and configure cluster service.
	tn := mux.Listen(muxMetaHeader)
	cs := cluster.NewService(tn, s)
	if err := cs.Open(); err != nil {
		log.Fatalf("failed to open cluster service: %s", err.Error())
	}

	// Create and configure HTTP service.
	h := httpd.New(config.C.CommonConf.HttpBind, cs)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	// If join was specified, make the join request.
	nodes, err := s.Nodes()
	if err != nil {
		log.Fatalf("get nodes failed: %s", err.Error())
	}

	// if exist a raftdb, and exist a cluster, don't join any leader.
	if joinAddr != "" && len(nodes) <= 1 {
		if err := join(joinAddr, c.ClusterBind); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	// update cluster meta
	if err := publishAPIAddr(cs, raftTn.Addr().String(), config.C.CommonConf.HttpBind, publishPeerTimeout); err != nil {
		log.Fatalf("failed to set peer for %s to %s: %s", raftTn.Addr().String(), config.C.CommonConf.HttpBind, err.Error())
	}
	log.Printf("set peer for %s to %s", raftTn.Addr().String(), config.C.CommonConf.HttpBind)

	log.Println("registry started successfully")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGKILL, syscall.SIGINT, syscall.SIGHUP, os.Interrupt)
	<-terminate

	// close HTTP service
	if err := h.Close(); err != nil {
		log.Println("close HTTP failed: %s", err)
	}

	// close cluster service
	if err := cs.Close(); err != nil {
		log.Println("close cluster service failed: %s", err)
	}

	// close store service
	if err := s.Close(true); err != nil {
		log.Println("close store failed: %s", err)
	}

	if err := os.Remove(config.C.CommonConf.PID); err != nil {
		log.Println("clean PID file failed: %s", err)
	}
	log.Println("registry exiting")
}

func join(joinAddr, raftAddr string) error {
	// Join using IP address, as that is what Hashicorp Raft works in.
	resv, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		return err
	}

	// Check for protocol scheme, and insert default if necessary.
	fullAddr := httpd.NormalizeAddr(fmt.Sprintf("%s/join", joinAddr))

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
			log.Println("join request redirecting to", fullAddr)
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
				log.Printf("failed to set peer for %s to %s: %s (retrying)",
					raftAddr, apiAddr, err.Error())
				continue
			}
			return nil
		case <-tmr.C:
			return fmt.Errorf("set peer timeout expired")
		}
	}
}
