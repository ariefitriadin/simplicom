package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	App struct {
		MyPodIP         string        `env:"MY_POD_IP"            envDefault:"127.0.0.1"`
		Domain          string        `env:"APP_DOMAIN"           envDefault:"http://localhost:3000"`
		Environment     string        `env:"APP_ENV"              envDefault:"local"`
		ShutdownTimeout time.Duration `env:"APP_SHUTDOWN_TIMEOUT" envDefault:"5s"`
		Secret          string        `env:"AUTH_SECRET"          envDefault:"secret"`
		AuthorizeURL    string        `env:"AUTH_AUTHORIZE_URL"   envDefault:"http://localhost:3000/authorize"`
	}
	HTTP struct {
		Host string `env:"HOST"      envDefault:"0.0.0.0"`
		Port int    `env:"HTTP_PORT" envDefault:"3000"`
		// Origins should follow format: scheme "://" host [ ":" port ]
		Origins []string `env:"HTTP_ORIGINS" envSeparator:"|" envDefault:"*"`

		ReadTimeout  time.Duration `env:"HTTP_SERVER_READ_TIMEOUT"     envDefault:"5s"`
		WriteTimeout time.Duration `env:"HTTP_SERVER_WRITE_TIMEOUT"    envDefault:"10s"`
		IdleTimeout  time.Duration `env:"HTTP_SERVER_SHUTDOWN_TIMEOUT" envDefault:"120s"`
	}
	GRPC struct {
		Host          string        `env:"HOST"      envDefault:"0.0.0.0"`
		Port          int           `env:"GRPC_PORT" envDefault:"3002"`
		ServerMinTime time.Duration `env:"GRPC_SERVER_MIN_TIME" envDefault:"5m"`  // if a client pings more than once every 5 minutes (default), terminate the connection
		ServerTime    time.Duration `env:"GRPC_SERVER_TIME"     envDefault:"2h"`  // ping the client if it is idle for 2 hours (default) to ensure the connection is still active
		ServerTimeout time.Duration `env:"GRPC_SERVER_TIMEOUT"  envDefault:"20s"` // wait 20 second (default) for the ping ack before assuming the connection is dead
		ConnTime      time.Duration `env:"GRPC_CONN_TIME"       envDefault:"10s"` // send pings every 10 seconds if there is no activity
		ConnTimeout   time.Duration `env:"GRPC_CONN_TIMEOUT"    envDefault:"20s"` // wait 20 second for ping ack before considering the connection dead
	}
	POSTGRES struct {
		Host     string `env:"PG_HOST"     envDefault:"0.0.0.0"`
		Port     int    `env:"PG_PORT"     envDefault:"5432"`
		User     string `env:"PG_USER"     envDefault:"user"`
		Pass     string `env:"PG_PASS"     envDefault:"secret"`
		Database string `env:"PG_DATABASE" envDefault:"productdb"`
	}
	Debug struct {
		Host string `env:"DEBUG_HOST" envDefault:"0.0.0.0"`
		Port int    `env:"DEBUG_PORT" envDefault:"4000"`
	}
}

func FromEnv() *Config {
	var c Config

	if err := env.Parse(&c.App); err != nil {
		panic(err)
	}
	if err := env.Parse(&c.GRPC); err != nil {
		panic(err)
	}
	if err := env.Parse(&c.HTTP); err != nil {
		panic(err)
	}
	if err := env.Parse(&c.POSTGRES); err != nil {
		panic(err)
	}
	if err := env.Parse(&c.Debug); err != nil {
		panic(err)
	}

	return &c
}
