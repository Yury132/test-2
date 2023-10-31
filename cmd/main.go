package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Yury132/Golang-Task-2/internal/config"
	service "github.com/Yury132/Golang-Task-2/internal/service/main_service"
	mediaService "github.com/Yury132/Golang-Task-2/internal/service/media_service"
	objectStorage "github.com/Yury132/Golang-Task-2/internal/storage/object-storage"
	"github.com/Yury132/Golang-Task-2/internal/storage/postgres"
	transport "github.com/Yury132/Golang-Task-2/internal/transport/http"
	"github.com/Yury132/Golang-Task-2/internal/transport/http/handlers"
	"github.com/Yury132/Golang-Task-2/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pressly/goose/v3"
)

const (
	dialect        = "pgx"
	commandUp      = "up"
	commandDown    = "down"
	migrationsPath = "./internal/migrations"
)

func main() {
	// Конфигурации
	cfg, err := config.Parse()
	if err != nil {
		panic(err)
	}

	// Логгер
	logger := cfg.Logger()

	// Миграции
	db, err := goose.OpenDBWithDriver(dialect, cfg.GetDBConnString())
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to open db by goose")
	}

	if err = goose.Run(commandUp, db, migrationsPath); err != nil {
		logger.Fatal().Msgf("migrate %v: %v", commandUp, err)
	}

	if err = db.Close(); err != nil {
		logger.Fatal().Err(err).Msg("failed to close db connection by goose")
	}

	// Настройка БД
	poolCfg, err := cfg.PgPoolConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to DB")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to db")
	}

	nc, err := nats.Connect(cfg.NATS.URL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to NATS")
	}
	defer func() {
		if err = nc.Drain(); err != nil {
			logger.Fatal().Err(err).Msg("failed to drain nats connection")
		}
	}()

	js, err := jetstream.New(nc)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create new jetstream")
	}

	streamCfg := jetstream.StreamConfig{
		Name:      "EVENTS",
		Retention: jetstream.WorkQueuePolicy,
		Subjects:  []string{"media.>"},
	}

	stream, err := js.CreateStream(ctx, streamCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create new stream")
	}

	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name: "media_service",
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create new consumer")
	}

	strg := postgres.New(conn)
	objStorage := objectStorage.New(logger)
	svc := service.New(logger, strg, objStorage, js)
	mediaSvc := mediaService.New(logger, strg, objStorage, cons)
	handler := handlers.New(logger, svc)
	server := transport.New(":8080").WithHandler(handler)

	wp := worker.New(logger, mediaSvc, 5)
	wp.Start()

	// graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT)

	go func() {
		logger.Info().Msg("Server starting...")
		if err = server.Run(); err != nil {
			logger.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	<-shutdown

	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func() {
		wp.Shutdown()

		defer wg.Done()
	}()
}
