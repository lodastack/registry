package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/cluster"
	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/httpd"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/registry/store"
	"github.com/lodastack/registry/tcp"
)

// Command line defaults
const (
	DefaultConfigFile = "/etc/registry/registry.conf"

	publishPeerDelay   = 1 * time.Second
	publishPeerTimeout = 60 * time.Second
	waitLeaderTimeout  = 10 * time.Second
)

// Command line parameters
var configFile string
var joinAddr string
var cpuProfile string
var memProfile string

// These variables are populated via the Go linker.
var (
	version   = "0"
	commit    = "unknown"
	branch    = "unknown"
	buildtime = "unknown"
)

const (
	muxRaftHeader = 1 // Raft consensus communications
	muxMetaHeader = 2 // Cluster meta communications
)

func init() {
	flag.StringVar(&configFile, "config", DefaultConfigFile, "Set the config file")
	flag.StringVar(&joinAddr, "join", "", "Set the leader API addr to join a cluster")
	flag.StringVar(&cpuProfile, "cpuprofile", "", "Write CPU profile to a file")
	flag.StringVar(&memProfile, "memprofile", "", "Write memory profile to a file")
}

// Main represents the program execution.
type Main struct {
	logger *log.Logger
}

// NewMain return a new instance of Main.
func NewMain() *Main {
	return &Main{
		logger: log.New(config.C.LogConf.Level, "main", model.LogBackend),
	}
}

func main() {
	flag.Parse()

	// Start requested profiling.
	startProfile(cpuProfile, memProfile)

	//parse config file
	err := config.ParseConfig(configFile)
	if err != nil {
		log.Errorf("Parse Config File Error : %s", err.Error())
		os.Exit(1)
	}

	// init log backend
	err = initLog(config.C.LogConf.Dir, config.C.LogConf.Level, config.C.LogConf.Logrotatenum, config.C.LogConf.Logrotatesize)
	if err != nil {
		log.Errorf("failed to new log backend: %s", err.Error())
		os.Exit(1)
	}

	m := NewMain()
	if err := m.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (m *Main) Start() error {

	m.logger.Printf("registry starting, version %s, branch %s, commit %s", version, branch, commit)

	//save pid to file
	err := ioutil.WriteFile(config.C.CommonConf.PID, []byte(strconv.Itoa(os.Getpid())), 0744)
	if err != nil {
		return fmt.Errorf("write PID file error: %s", err.Error())
	}

	// store config
	c := config.C.DataConf

	// serve mux TCP
	ln, err := net.Listen("tcp", c.ClusterBind)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %s", c.ClusterBind, err.Error())
	}
	mux := tcp.NewMux(ln, nil)
	go mux.Serve()

	// Start up mux and get transports for cluster.
	raftTn := mux.Listen(muxRaftHeader)
	s := store.New(c.Dir, raftTn)
	if err := s.Open(joinAddr == ""); err != nil {
		return fmt.Errorf("failed to open store: %s", err.Error())
	}

	// Create and configure cluster service.
	tn := mux.Listen(muxMetaHeader)
	cs := cluster.NewService(tn, s)
	if err := cs.Open(); err != nil {
		return fmt.Errorf("failed to open cluster service: %s", err.Error())
	}

	// If join was specified, make the join request.
	nodes, err := s.Nodes()
	if err != nil {
		return fmt.Errorf("get nodes failed: %s", err.Error())
	}

	// if exist a raftdb, or exist a cluster, don't join any leader.
	if joinAddr != "" && len(nodes) <= 1 {
		if err := join(joinAddr, c.ClusterBind); err != nil {
			return fmt.Errorf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	// wait for leader
	l, err := s.WaitForLeader(waitLeaderTimeout)
	if err != nil || l == "" {
		return fmt.Errorf("wait leader failed: %s", err.Error())
	}
	m.logger.Printf("cluster leader is: %s", l)

	// update cluster meta
	if err := publishAPIAddr(cs, raftTn.Addr().String(), config.C.CommonConf.HttpBind, publishPeerTimeout); err != nil {
		return fmt.Errorf("failed to set peer for %s to %s: %s", raftTn.Addr().String(), config.C.CommonConf.HttpBind, err.Error())
	}
	m.logger.Printf("set peer for %s to %s", raftTn.Addr().String(), config.C.CommonConf.HttpBind)

	// Create and configure HTTP service.
	h, err := httpd.New(config.C.HTTPConf, cs)
	if err := h.Start(); err != nil {
		return fmt.Errorf("failed to start HTTP service: %s", err.Error())
	}

	m.logger.Printf("registry started successfully")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL)
	<-terminate

	// close HTTP service
	if err := h.Close(); err != nil {
		m.logger.Errorf("close HTTP failed: %s", err)
	}

	// close cluster service
	if err := cs.Close(); err != nil {
		m.logger.Errorf("close cluster service failed: %s", err)
	}

	// close store service
	if err := s.Close(true); err != nil {
		m.logger.Errorf("close store failed: %s", err)
	}

	if err := os.Remove(config.C.CommonConf.PID); err != nil {
		m.logger.Errorf("clean PID file failed: %s", err)
	}
	stopProfile()

	// flush log
	model.LogBackend.Flush()

	m.logger.Printf("registry exiting")
	return nil
}
