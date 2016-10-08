package config

import (
	"sync"

	"github.com/BurntSushi/toml"
)

const (
	//APP NAME
	AppName = "Registry"
	//Usage
	Usage = "Registry Usage"
	//Vresion Num
	Version = "0.0.1"
	//Author Nmae
	Author = "LoadStack Developer Group"
	//Email Address
	Email = "oiooj@qq.com"
)

var (
	mux sync.RWMutex

	// global config
	C Config
)

type Config struct {
	CommonConf CommonConfig `toml:"common"`
	DataConf   DataConfig   `toml:"data"`
	LogConf    LogConfig    `toml:"log"`
}

type CommonConfig struct {
	HttpBind string `toml:"httpbind"`
	PID      string `toml:"pid"`
}

type DataConfig struct {
	Dir           string `toml:"dir"`
	ClusterBind   string `toml:"clusterbind"`
	ClusterLeader string `toml:"clusterleader"`
}

// LogConfig is log config struct
type LogConfig struct {
	Dir           string `toml:"logdir"`
	Level         string `toml:"loglevel"`
	Logrotatenum  int    `toml:"logrotatenum"`
	Logrotatesize uint64 `toml:"logrotatesize"`
}

func ParseConfig(path string) error {
	mux.Lock()
	defer mux.Unlock()

	if _, err := toml.DecodeFile(path, &C); err != nil {
		return err
	}
	return nil
}

func GetConfig() Config {
	mux.RLock()
	defer mux.RUnlock()
	return C
}
