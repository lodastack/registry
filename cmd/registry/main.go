package main // "import github.com/lodastack/registry"

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/lodastack/log"
	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/dns"
	"github.com/lodastack/registry/httpd"
	"github.com/lodastack/registry/model"
	"github.com/lodastack/store/cluster"
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

// Start starts main registry service
func (m *Main) Start() error {

	m.logger.Printf("registry starting, version %s, branch %s, commit %s", version, branch, commit)

	//save pid to file
	err := ioutil.WriteFile(config.C.CommonConf.PID, []byte(strconv.Itoa(os.Getpid())), 0744)
	if err != nil {
		return fmt.Errorf("write PID file error: %s", err.Error())
	}

	// store config
	c := config.C.DataConf

	storeLogger := log.New(config.C.LogConf.Level, "store", model.LogBackend)
	opts := cluster.Options{
		Bind:     c.ClusterBind,
		DataDir:  c.Dir,
		JoinAddr: joinAddr,
		Logger:   storeLogger,
	}
	cs, err := cluster.NewService(opts)
	if err != nil {
		return fmt.Errorf("new store service failed: %s", err.Error())
	}

	if err := cs.Open(); err != nil {
		return fmt.Errorf("failed to open cluster service failed: %s", err.Error())
	}

	// If join was specified, make the join request.
	nodes, err := cs.Nodes()
	if err != nil {
		return fmt.Errorf("get nodes failed: %s", err.Error())
	}

	// if exist a raftdb, or exist a cluster, don't join any leader.
	if joinAddr != "" && len(nodes) <= 1 {
		if err := cs.JoinCluster(joinAddr, c.ClusterBind); err != nil {
			return fmt.Errorf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	// wait for leader
	l, err := cs.WaitForLeader(waitLeaderTimeout)
	if err != nil || l == "" {
		return fmt.Errorf("wait leader failed: %s", err.Error())
	}
	m.logger.Printf("cluster leader is: %s", l)

	// update cluster meta
	if err := cs.PublishAPIAddr(config.C.HTTPConf.Bind, publishPeerDelay, publishPeerTimeout); err != nil {
		return fmt.Errorf("failed to set peer to [API:%s]: %s", config.C.HTTPConf.Bind, err.Error())
	}

	// Create and configure HTTP service.
	h, err := httpd.New(config.C.HTTPConf, cs)
	if err := h.Start(); err != nil {
		return fmt.Errorf("failed to start HTTP service: %s", err.Error())
	}

	// DNS service
	dns, err := dns.New(config.C.DNSConf, cs)
	if err := dns.Start(); err != nil {
		return fmt.Errorf("failed to start DNS service: %s", err.Error())
	}

	m.logger.Printf("registry started successfully")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL)
	<-terminate
	stopProfile()

	// close DNS service
	if err := dns.Close(); err != nil {
		m.logger.Errorf("close DNS failed: %s", err)
	}

	// close HTTP service
	if err := h.Close(); err != nil {
		m.logger.Errorf("close HTTP failed: %s", err)
	}

	// close cluster service
	if err := cs.Close(); err != nil {
		m.logger.Errorf("close cluster service failed: %s", err)
	}

	if err := os.Remove(config.C.CommonConf.PID); err != nil {
		m.logger.Errorf("clean PID file failed: %s", err)
	}

	// flush log
	model.LogBackend.Flush()

	m.logger.Printf("registry exiting")
	return nil
}
