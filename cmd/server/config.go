package main

import (
	"time"

	"github.com/caarlos0/env/v6"
	zlog "github.com/rs/zerolog/log"
)

type Config struct {
	DB         DB
	RESTServer RESTServer
	AMQP       AMQP
}

type DB struct {
	PostgresDSN string `env:"DB_POSTGRES_DSN,required" envDefault:"host=localhost port=5432 user=user password=password dbname=rest-auth-example sslmode=disable"`

	MaxOpenConnections    int           `env:"DB_MAX_OPEN_CONNECTIONS" envDefault:"10"`
	MaxIdleConnections    int           `env:"DB_MAX_IDLE_CONNECTIONS" envDefault:"5"`
	MaxConnectionLifetime time.Duration `env:"DB_MAX_CONNECTION_LIFETIME" envDefault:"5m"`
}

type RESTServer struct {
	Address string `env:"SERVER_ADDRESS" envDefault:"0.0.0.0:9000"`

	Timeout time.Duration `env:"SERVER_TIMEOUT" envDefault:"10s"`

	JWTSecret      string        `env:"JWT_SECRET,required" envDefault:"JWT_SECRET"`
	AccessTokenTTL time.Duration `env:"ACCESS_TOKEN_TTL" envDefault:"1h"`
}

type AMQP struct {
	ConnectionURL string `env:"AMQP_CONNECTION_URL" envDefault:"amqp://user:pass@localhost"`

	ExchangeName string `env:"AMQP_EXCHANGE_NAME" envDefault:"rest_auth_example_tasks"`
	RoutingKey   string `env:"AMQP_ROUTING_KEY" envDefault:"stats_task"`
}

func ReadConfig() Config {
	var conf Config
	err := env.Parse(&conf)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to read the config")
	}

	return conf
}
