package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"runtime"
	"time"
)

var redisURL = "redis://localhost:6379"

var (
	validTimeoutUnitOptions = map[string]time.Duration{
		"second":      time.Second,
		"millisecond": time.Millisecond,
	}
	validLoggingFormatOptions = map[string]bool{
		"text": true,
		"json": true,
	}
	validStorageOptions = map[string]bool{
		"redis": true,
	}
	validJobQueueOptions = map[string]bool{
		"redis": true,
	}
	validProtocolOptions = map[string]bool{
		"http": true,
	}
)

type HTTP struct {
	Port string `yaml:"port"`
}

type Server struct {
	Protocol string `yaml:"protocol"`
	HTTP     HTTP   `yaml:"http"`
}

type QueueParams struct {
	Name              string `yaml:"name"`
	Durable           bool   `yaml:"durable"`
	DeletedWhenUnused bool   `yaml:"deleted_when_unused"`
	Exclusive         bool   `yaml:"exclusive"`
	NoWait            bool   `yaml:"no_wait"`
}

type ConsumeParams struct {
	Name      string `yaml:"name"`
	AutoACK   bool   `yaml:"auto_ack"`
	Exclusive bool   `yaml:"exclusive"`
	NoLocal   bool   `yaml:"no_local"`
	NoWait    bool   `yaml:"no_wait"`
}

type PublishParams struct {
	Exchange   string `yaml:"exchange"`
	RoutingKey string `yaml:"routing_key"`
	Mandatory  bool   `yaml:"mandatory"`
	Immediate  bool   `yaml:"immediate"`
}

type JobQueue struct {
	Option string `yaml:"option"`
	Redis  Redis  `yaml:"redis"`
}

type WorkerPool struct {
	Workers       int `yaml:"workers"`
	QueueCapacity int `yaml:"queue_capacity"`
}

type Scheduler struct {
	StoragePollingInterval  int `yaml:"storage_polling_interval"`
	JobQueuePollingInterval int `yaml:"job_queue_polling_interval"`
}

type Storage struct {
	Option string `yaml:"option"`
	Redis  Redis  `yaml:"redis"`
}

type Redis struct {
	URL          string
	KeyPrefix    string `yaml:"key_prefix"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

type Config struct {
	Server            Server     `yaml:"server"`
	JobQueue          JobQueue   `yaml:"job_queue"`
	WorkerPool        WorkerPool `yaml:"worker_pool"`
	Scheduler         Scheduler  `yaml:"scheduler"`
	Storage           Storage    `yaml:"storage"`
	TimeoutUnitOption string     `yaml:"timeout_unit"`
	LoggingFormat     string     `yaml:"logging_format"`
	TimeoutUnit       time.Duration
}

func (cfg *Config) Load(filepath string) error {
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return err
	}
	err = cfg.setServerConfig()
	if err != nil {
		return err
	}
	err = cfg.setJobQueueConfig()
	if err != nil {
		return err
	}
	cfg.setWorkerPoolConfig()
	err = cfg.setTimeoutUnitConfig()
	if err != nil {
		return err
	}
	cfg.setSchedulerConfig()
	err = cfg.setLoggingFormatConfig()
	if err != nil {
		return err
	}
	err = cfg.setStorageConfig()
	if err != nil {
		return err
	}
	return nil
}

func (cfg *Config) setServerConfig() error {
	if _, ok := validProtocolOptions[cfg.Server.Protocol]; !ok {
		return fmt.Errorf("%s is not a valid protocol option, valid options: %v", cfg.Server.Protocol, validProtocolOptions)
	}
	if cfg.Server.HTTP.Port == "" {
		cfg.Server.HTTP.Port = "8080"
	}

	return nil
}

func (cfg *Config) setJobQueueConfig() error {
	if _, ok := validJobQueueOptions[cfg.JobQueue.Option]; !ok {
		return fmt.Errorf("%s is not a valid job queue option, valid options: %v", cfg.JobQueue.Option, validJobQueueOptions)
	}
	if cfg.JobQueue.Option == "redis" {
		url := redisURL
		cfg.JobQueue.Redis.URL = url
	}
	return nil
}

func (cfg *Config) setWorkerPoolConfig() {
	if cfg.WorkerPool.Workers == 0 {
		// Defaults to number of cores.
		cfg.WorkerPool.Workers = runtime.NumCPU()
	}
	if cfg.WorkerPool.QueueCapacity == 0 {
		cfg.WorkerPool.QueueCapacity = cfg.WorkerPool.Workers * 2
	}
}

func (cfg *Config) setTimeoutUnitConfig() error {
	timeoutUnit, ok := validTimeoutUnitOptions[cfg.TimeoutUnitOption]
	if !ok {
		return fmt.Errorf("%s is not a valid timeout_unit option, valid options: %v", cfg.TimeoutUnitOption, validTimeoutUnitOptions)
	}
	cfg.TimeoutUnit = timeoutUnit
	return nil
}

func (cfg *Config) setSchedulerConfig() {
	if cfg.Scheduler.StoragePollingInterval == 0 {
		if cfg.TimeoutUnit == time.Second {
			cfg.Scheduler.StoragePollingInterval = 60
			cfg.Scheduler.JobQueuePollingInterval = 1
		} else {
			cfg.Scheduler.StoragePollingInterval = 60000
			cfg.Scheduler.JobQueuePollingInterval = 1000
		}
	}
}

func (cfg *Config) setLoggingFormatConfig() error {
	if _, ok := validLoggingFormatOptions[cfg.LoggingFormat]; !ok {
		return fmt.Errorf("%s is not a valid logging_format option, valid options: %v", cfg.LoggingFormat, validLoggingFormatOptions)
	}
	return nil
}

func (cfg *Config) setStorageConfig() error {
	if _, ok := validStorageOptions[cfg.Storage.Option]; !ok {
		return fmt.Errorf("%s is not a valid storage option, valid options: %v", cfg.Storage.Option, validStorageOptions)
	}
	if cfg.Storage.Option == "redis" {
		url := redisURL
		//if url == "" {
		//	return errors.New("Redis URL not provided")
		//}
		cfg.Storage.Redis.URL = url

		if cfg.Storage.Redis.PoolSize == 0 {
			cfg.Storage.Redis.PoolSize = 10
		}
		if cfg.Storage.Redis.MinIdleConns == 0 {
			cfg.Storage.Redis.MinIdleConns = 10
		}
	}
	return nil
}
