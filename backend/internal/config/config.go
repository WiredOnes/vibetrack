package config

import (
	"time"
)

type Service struct {
	Name        string `yaml:"name" validate:"required"`
	Environment string `yaml:"environment" validate:"oneof=testing development staging production"`
}

type Logger struct {
	Developement bool   `yaml:"development"`
	StdoutOnly   bool   `yaml:"stdout_only"`
	OutputFile   string `yaml:"output_file" validate:"required"`

	MaxFileSizeMB  int  `yaml:"max_file_size_mb" validate:"gte=0"`
	MaxFilesAmount int  `yaml:"max_backups_amount" validate:"gte=1"`
	MaxFileAgeDays int  `yaml:"max_file_age_days" validate:"gte=1"`
	Compression    bool `yaml:"compression"`

	TimeKey       string `yaml:"time_key"`
	LevelKey      string `yaml:"level_key"`
	NameKey       string `yaml:"name_key"`
	CallerKey     string `yaml:"caller_key"`
	MessageKey    string `yaml:"message_key"`
	StacktraceKey string `yaml:"stacktrace_key"`
}

type GRPC struct {
	Host string `yaml:"host" validate:"ip"`
	Port uint   `yaml:"port" validate:"port"`

	EnableTLS bool `yaml:"enable_tls"`

	EnableHealth     bool `yaml:"enable_health"`
	EnableReflection bool `yaml:"enable_reflection"`

	KeepaliveMaxConnectionIdle     time.Duration `yaml:"keepalive_max_connection_idle" validate:"min=1s"`
	KeepaliveMaxConnectionAge      time.Duration `yaml:"keepalive_max_connection_age" validate:"min=1s"`
	KeepaliveMaxConnectionAgeGrace time.Duration `yaml:"keepalive_max_connection_age_grace" validate:"min=1s"`
	KeepaliveTime                  time.Duration `yaml:"keepalive_time" validate:"min=1s"`
	KeepaliveTimeout               time.Duration `yaml:"keepalive_timeout" validate:"min=1s"`

	KeepaliveMinTime             time.Duration `yaml:"keepalive_min_time" validate:"min=1s"`
	KeepalivePermitWithoutStream bool          `yaml:"keepalive_permit_without_stream"`

	MaxReceiveMessageSizeMB int `yaml:"max_receive_message_size_mb" validate:"gte=4"`
	MaxSendMessageSizeMB    int `yaml:"max_send_message_size_mb" validate:"gte=4"`

	Timeout time.Duration `yaml:"timeout" validate:"min=1s"`
}

type HTTP struct {
	Host string `yaml:"host" validate:"ip"`
	Port uint   `yaml:"port" validate:"port"`

	EnableTLS bool `yaml:"enable_tls"`

	ReadTimeout  time.Duration `yaml:"read_timeout" validate:"min=1s"`
	WriteTimeout time.Duration `yaml:"write_timeout" validate:"min=1s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" validate:"min=1s"`
}

type DB struct {
	Host string `yaml:"host" validate:"required"`
	Port uint   `yaml:"port" validate:"port"`

	User     string `yaml:"user" validate:"required"`
	Password string `yaml:"password" validate:"required"`
	Database string `yaml:"database" validate:"required"`
	SSLMode  string `yaml:"ssl_mode" validate:"oneof=disable require verify-ca verify-full"`

	MigrationsTableName string `yaml:"migrations_table_name" validate:"required"`
	// TODO: Connection pool settings
}

type Job struct {
	Name         string        `yaml:"name" validate:"required"`
	RunOnStartup bool          `yaml:"run_on_startup"`
	Period       time.Duration `yaml:"period" validate:"gte=0"`
	Timeout      time.Duration `yaml:"timeout" validate:"gte=0"`
}

type GaugeMetric struct {
	Name   string   `yaml:"name" validate:"required"`
	Labels []string `yaml:"labels"`
}

type CounterMetric struct {
	Name   string   `yaml:"name" validate:"required"`
	Labels []string `yaml:"labels"`
}

type HistogramMetric struct {
	Name    string    `yaml:"name" validate:"required"`
	Labels  []string  `yaml:"labels"`
	Buckets []float64 `yaml:"buckets"`
}

type SummaryMetric struct {
	Name       string              `yaml:"name" validate:"required"`
	Labels     []string            `yaml:"labels"`
	Objectives map[float64]float64 `yaml:"buckets"`
	MaxAge     time.Duration       `yaml:"max_age" validate:"min=1s"`
	AgeBuckets uint32              `yaml:"age_buckets" validate:"gte=1"`
}

type Monitoring struct {
	Gauges     []GaugeMetric     `yaml:"gauges"`
	Counters   []CounterMetric   `yaml:"counters"`
	Histograms []HistogramMetric `yaml:"histograms"`
	Summaries  []SummaryMetric   `yaml:"summaries"`
}

type Config struct {
	Service    Service    `yaml:"service"`
	Logger     Logger     `yaml:"logger"`
	GRPC       GRPC       `yaml:"grpc"`
	HTTP       HTTP       `yaml:"http"`
	DB         DB         `yaml:"db"`
	Monitoring Monitoring `yaml:"monitoring"`
	Jobs       []Job      `yaml:"jobs"`
}
