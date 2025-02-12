package config

import (
	"fmt"
	"time"
)

type APIServerHTTPConfig struct {
	Host              string        `env:"HOST,default=127.0.0.1"`
	Port              uint16        `env:"PORT,default=8080"`
	ReadTimeout       time.Duration `env:"READ_TIMEOUT,default=30s"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT,default=30s"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT,default=30s"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT,default=30s"`
	MaxBodyBytes      int64         `env:"MAX_BODY_BYTES,default=1048576"`
}

func (a APIServerHTTPConfig) Address() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

type HTTPClientConfig struct {
	Scheme        string        `env:"SCHEME,default=http"`
	Host          string        `env:"HOST,default=127.0.0.1"`
	Port          uint16        `env:"PORT,default=8080"`
	Timeout       time.Duration `env:"TIMEOUT,default=30s"`
	MaxRetries    int           `env:"MAX_RETRIES,default=3"`
	RetryInterval time.Duration `env:"RETRY_INTERVAL,default=2s"`
	RateLimit     struct {
		Burst int     `env:"BURST,default=0"`
		RPS   float64 `env:"RPS,default=20.0"`
	} `env:",prefix=RATE_LIMIT_"`
}

func (c HTTPClientConfig) Address() string {
	return fmt.Sprintf("%s://%s:%d", c.Scheme, c.Host, c.Port)
}

type ObservabilityHTTPConfig struct {
	Host         string        `env:"HOST,default=127.0.0.1"`
	Port         uint16        `env:"PORT,default=8383"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT,default=30s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT,default=30s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT,default=1m"`
}

func (a ObservabilityHTTPConfig) Address() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}
