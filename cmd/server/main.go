package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lodthe/rest-auth-example/internal/muser"
	"github.com/lodthe/rest-auth-example/internal/statstask"
	"github.com/lodthe/rest-auth-example/internal/taskqueue"
	"github.com/lodthe/rest-auth-example/pkg/restapi"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/wagslane/go-rabbitmq"
)

func main() {
	conf := ReadConfig()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	db, err := setupDatabaseConnection(conf.DB)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to setup database connection")
	}
	defer db.Close()

	publisher, err := rabbitmq.NewPublisher(
		conf.AMQP.ConnectionURL,
		rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to connect to RabbitMQ")
	}
	defer publisher.Close()

	taskRepo := statstask.NewRepository(db)
	userRepo := muser.NewRepository(db)
	producer := taskqueue.NewProducer(publisher, conf.AMQP.ExchangeName, conf.AMQP.RoutingKey)

	router := restapi.NewRouter(userRepo, taskRepo, producer, conf.RESTServer.Timeout)

	zlog.Info().Str("address", conf.RESTServer.Address).Msg("starting the server...")

	srv := &http.Server{
		Addr:    conf.RESTServer.Address,
		Handler: router,
	}
	go func() {
		err = srv.ListenAndServe()
		if err != nil {
			zlog.Fatal().Err(err).Msg("listen failed")
		}
	}()

	<-stop
	cancel()

	shutdownCtx, shutdown := context.WithTimeout(ctx, 10*time.Second)
	defer shutdown()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		zlog.Error().Err(err).Msg("server shutdown failed")
	}
}

func setupDatabaseConnection(config DB) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", config.PostgresDSN)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(config.MaxConnectionLifetime)
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)

	return db, nil
}
