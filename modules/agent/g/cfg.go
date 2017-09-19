// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package g

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"os/exec"
	"bytes"
	"github.com/toolkits/file"
	"github.com/garyburd/redigo/redis"
)

type PluginConfig struct {
	Enabled bool   `json:"enabled"`
	Dir     string `json:"dir"`
	Git     string `json:"git"`
	LogDir  string `json:"logs"`
}

type HeartbeatConfig struct {
	Enabled  bool   `json:"enabled"`
	Addr     string `json:"addr"`
	Interval int    `json:"interval"`
	Timeout  int    `json:"timeout"`
}

type TransferConfig struct {
	Enabled  bool     `json:"enabled"`
	Addrs    []string `json:"addrs"`
	Interval int      `json:"interval"`
	Timeout  int      `json:"timeout"`
}

type HttpConfig struct {
	Enabled  bool   `json:"enabled"`
	Listen   string `json:"listen"`
	Backdoor bool   `json:"backdoor"`
}

type CollectorConfig struct {
	IfacePrefix []string `json:"ifacePrefix"`
	MountPoint  []string `json:"mountPoint"`
}

type CmdbRedisConfig struct {
	Enabled  bool `json:"enabled"`
	Addr     string `json:"addr"`
	Password string `json:"password"`
}

type GlobalConfig struct {
	Debug         bool              `json:"debug"`
	Hostname      string            `json:"hostname"`
	IP            string            `json:"ip"`
	Plugin        *PluginConfig     `json:"plugin"`
	Heartbeat     *HeartbeatConfig  `json:"heartbeat"`
	Transfer      *TransferConfig   `json:"transfer"`
	Http          *HttpConfig       `json:"http"`
	CmdbRedis     *CmdbRedisConfig  `json:"redis"`
	Collector     *CollectorConfig  `json:"collector"`
	DefaultTags   map[string]string `json:"default_tags"`
	IgnoreMetrics map[string]bool   `json:"ignore"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	lock       = new(sync.RWMutex)
)

type HostInfo struct {
	CoName      string
	IP          string
	HostName    string
	NetworkType string
	VpcId       int
	DiskType    string
	ID          int
}

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}



func Sn() (string, error) {
	cmd := exec.Command("/bin/bash", "-c", "dmidecode -s system-serial-number|sed 's/ //g'")
	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Println("ERROR: execute dmidecode to get sn failed")
	}
	sn := out.String()
	return sn, nil
}

func RedisHostName(sn string) (string, error) {
	redis_addr := Config().CmdbRedis.Addr
	redis_pass := Config().CmdbRedis.Password
	log.Println("INFO: redis_addr:", redis_addr, ",redis_pass:", redis_pass)
	if redis_addr == "" {
		log.Println("ERROR: read redis_addr failed")
	}

	c, err := redis.Dial("tcp", redis_addr)
	if err != nil {
		log.Println("ERROR: can't connect redis, redis_addr: ", redis_addr)
	}
	c.Do("AUTH", redis_pass)
	value, err := redis.String(c.Do("GET", sn))
	if err != nil {
		log.Println("ERROR: can't get hostinfo from redis which the machine'sn is ",sn)
	}
	var host_info HostInfo
	json.Unmarshal([]byte(value), &host_info)
	hostname := host_info.HostName
	log.Println("INFO: host name from redis is ", hostname)
	defer c.Close()
	return hostname, nil
}

func Hostname() (string, error) {
	redisenabled := Config().CmdbRedis.Enabled
	log.Println("INFO: redis enabled is ", redisenabled)
	if redisenabled == true {
		var sn string
		var hostname string
		sn, _ = Sn()
		hostname, _ = RedisHostName(sn)
		if hostname != "" {
			return hostname, nil
		}
	}
	hostname := Config().Hostname
	if hostname != "" {
		return hostname, nil
	}

	if os.Getenv("FALCON_ENDPOINT") != "" {
		hostname = os.Getenv("FALCON_ENDPOINT")
		return hostname, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("ERROR: os.Hostname() fail", err)
	}
	return hostname, err
}

func IP() string {
	ip := Config().IP
	if ip != "" {
		// use ip in configuration
		return ip
	}

	if len(LocalIp) > 0 {
		ip = LocalIp
	}

	return ip
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	lock.Lock()
	defer lock.Unlock()

	config = &c

	log.Println("read config file:", cfg, "successfully")
	log.Println("New log record:", cfg, "successfully")
}
