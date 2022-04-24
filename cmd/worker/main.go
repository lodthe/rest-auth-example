package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lodthe/rest-auth-example/internal/muser"
	"github.com/lodthe/rest-auth-example/internal/statstask"
	"github.com/lodthe/rest-auth-example/internal/taskqueue"
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

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID && region == "ru-central1" {
			return aws.Endpoint{
				PartitionID:   "yc",
				URL:           "https://storage.yandexcloud.net",
				SigningRegion: "conf.S3.Region",
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})

	cfg, err := config.LoadDefaultConfig(ctx, config.WithDefaultRegion(conf.S3.Region), config.WithEndpointResolverWithOptions(customResolver))
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.NewFromConfig(cfg)
	taskRepo := statstask.NewRepository(db)
	userRepo := muser.NewRepository(db)
	worker := statstask.NewWorker(ctx, taskRepo, userRepo, s3Client, conf.S3.Bucket)

	rabbitConsumer, err := rabbitmq.NewConsumer(
		conf.AMQP.ConnectionURL,
		rabbitmq.Config{},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConsumer.Close()

	consumer := taskqueue.NewConsumer(rabbitConsumer, conf.AMQP.QueueName, conf.AMQP.RoutingKey)

	go func() {
		err := consumer.StartConsuming(worker.HandleTask)
		if err != nil {
			zlog.Fatal().Err(err).Msg("consumer failed")
		}
	}()

	zlog.Info().Msg("consumer has been started")

	<-stop
	cancel()
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
