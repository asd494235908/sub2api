package config

type ArchiveConfig struct {
	Enabled              bool               `mapstructure:"enabled"`
	QueueCapacity        int                `mapstructure:"queue_capacity"`
	WorkerCount          int                `mapstructure:"worker_count"`
	BatchSize            int                `mapstructure:"batch_size"`
	FlushIntervalSeconds int                `mapstructure:"flush_interval_seconds"`
	OverflowPolicy       string             `mapstructure:"overflow_policy"`
	InlineDataMaxBytes   int                `mapstructure:"inline_data_max_bytes"`
	MinIO                ArchiveMinIOConfig `mapstructure:"minio"`
}

type ArchiveMinIOConfig struct {
	Endpoint       string `mapstructure:"endpoint"`
	Bucket         string `mapstructure:"bucket"`
	AccessKey      string `mapstructure:"access_key"`
	SecretKey      string `mapstructure:"secret_key"`
	Region         string `mapstructure:"region"`
	ForcePathStyle bool   `mapstructure:"force_path_style"`
}

const (
	PromptArchiveOverflowPolicyDropAndLog = "drop_and_log"
)
