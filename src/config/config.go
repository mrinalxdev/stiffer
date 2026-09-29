package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	StageIngest = "ingest"
	StageAnalyze = "analyze"
	StageProxy = "proxy"
	StageIndex = "index"
	StageAll = "all"
)

var ValidStages = []string{StageIngest, StageAnalyze, StageProxy, StageIndex}

type Config struct {
	NATSURL string
	StreamName string
	SubjectRoot string
	Stage string
	Concurrency map[string]int
	AckWait    time.Duration
	MaxDeliver int
	NakDelay   time.Duration
	Duplicates time.Duration
	MaxAge     time.Duration
	Replicas   int
}

func Load() (*Config, error) {
	c := &Config{
		NATSURL:     env("NATS_URL", "nats://127.0.0.1:4222"),
		StreamName:  env("NATS_STREAM", "MAM"),
		SubjectRoot: env("NATS_SUBJECT_ROOT", "mam"),
		Stage:       strings.ToLower(env("STAGE", StageAll)),
		AckWait:     envDuration("ACK_WAIT", 2*time.Minute),
		MaxDeliver:  envInt("MAX_DELIVER", 5),
		NakDelay:    envDuration("NAK_DELAY", 5*time.Second),
		Duplicates:  envDuration("DEDUP_WINDOW", 2*time.Minute),
		MaxAge:      envDuration("STREAM_MAX_AGE", 24*time.Hour),
		Replicas:    envInt("STREAM_REPLICAS", 1),
		Concurrency: map[string]int{
			StageIngest:  envInt("CONCURRENCY_INGEST", 8),
			StageAnalyze: envInt("CONCURRENCY_ANALYZE", 2),
			StageProxy:   envInt("CONCURRENCY_PROXY", 4),
			StageIndex:   envInt("CONCURRENCY_INDEX", 8),
		},
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) validate() error {
	if c.Stage != StageAll {
		ok := false
		for _, s := range ValidStages {
			if c.Stage == s {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("unknown STAGE %q (want all|ingest|analyze|proxy|index)", c.Stage)
		}
	}
	if c.SubjectRoot == "" || strings.Contains(c.SubjectRoot, ".") {
		return fmt.Errorf("NATS_SUBJECT_ROOT must be a single token, got %q", c.SubjectRoot)
	}
	return nil
}

func (c *Config) StagesToRun() []string {
	if c.Stage == StageAll {
		return append([]string{}, ValidStages...)
	}
	return []string{c.Stage}
}

func env(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}

func envDuration(key string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
