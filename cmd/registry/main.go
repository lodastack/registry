package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/lodastack/registry/config"
	"github.com/lodastack/registry/httpd"
	"github.com/lodastack/registry/store"
)

// Command line defaults
const (
	DefaultConfigFile = "/etc/registry/registry.conf"
)

// Command line parameters
var configFile string

func init() {
	flag.StringVar(&configFile, "config", DefaultConfigFile, "Set the config file")
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
	joinAddr := c.ClusterLeader

	s := store.New(c.Dir, c.ClusterBind)
	if err := s.Open(joinAddr == ""); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	h := httpd.New(config.C.CommonConf.HttpBind, s)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to start HTTP service: %s", err.Error())
	}

	// If join was specified, make the join request.
	if joinAddr != "" {
		if err := join(joinAddr, c.ClusterBind); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	log.Println("registry started successfully")

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate

	// close store service
	if err := s.Close(true); err != nil {
		log.Println("close store failed: %s", err)
	}

	// close HTTP service
	if err := h.Close(); err != nil {
		log.Println("close HTTP failed: %s", err)
	}

	if err := os.Remove(config.C.CommonConf.PID); err != nil {
		log.Println("clean PID file failed: %s", err)
	}
	log.Println("registry exiting")
}

func join(joinAddr, raftAddr string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
