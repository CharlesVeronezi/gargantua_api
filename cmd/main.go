package main

import (
	"context"
	"errors"
	"fmt"
	"gargantua/internal/infra/httpapi"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/rueidis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)
	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Println(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println("Finalized API")
}

func run(ctx context.Context) error {
	// Configuração do logger
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	logger = logger.Named("gargantua")
	defer func() { _ = logger.Sync() }()

	// Conexão com MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://db:27017"))
	if err != nil {
		return err
	}

	// Desconecta o cliente MongoDB no final
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("failed to disconnect MongoDB client", zap.Error(err))
		}
	}()

	// Verifica a conexão com o MongoDB
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

		//Conexao do cliente redis
		rdClient, err := rueidis.NewClient(rueidis.ClientOption{InitAddress: []string{"cache:6379"}})
		if err != nil {
			return err
		}
		rdClient.B().Ping()

	// Configuração da API e do roteador
	si := httpapi.NewAPI(client, logger, rdClient)
	r := chi.NewMux()
	r.Use(middleware.RequestID, middleware.Recoverer)
	r.Mount("/", httpapi.Handler(si))

	// Configuração do servidor HTTP
	srv := &http.Server{
		//Addr:         ":8080",
		Addr:         ":" + os.Getenv("HTTP_PORT"),
		Handler:      r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Desliga o servidor
	defer func() {
		const timeout = 30 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("failed to shutdown server", zap.Error(err))
		}
	}()

	// Inicia o servidor em uma goroutine
	errChan := make(chan error, 1)
	go func() {
		//println("Server started successfully || Port: 8080")
		println("Server started successfully || Port:", os.Getenv("HTTP_PORT"))
		if err := srv.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	// Aguarda sinal de cancelamento ou erro do servidor
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}
