package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"production-go-api-template/api/resource"
	"production-go-api-template/api/router"
	"production-go-api-template/api/router/middleware"
	"production-go-api-template/config"
	"production-go-api-template/pkg/logger"
	"syscall"

	"github.com/rs/zerolog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	c, err := config.New()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	lvl := zerolog.InfoLevel
	var logLevel gormlogger.LogLevel
	if c.Server.Debug {
		lvl = zerolog.TraceLevel
		logLevel = gormlogger.Info
	}
	l := logger.New(lvl)

	db := openDatabase(c.DB.DBPath, l, logLevel)
	migrate(db, l)

	mux := router.SetupRouter(db)

	stack := middleware.CreateStack(
		middleware.RequestID,
		middleware.InjectDeps(c, l),
		middleware.CORS(c.Server.CorsOrigins),
		middleware.ContentTypeJSON,
		middleware.RequestLog(l),
	)

	authenticator := middleware.NewAuthenticator(c.Auth, c.Security, l).Middleware()

	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/livez" {
			stack(mux).ServeHTTP(w, r)
		} else {
			stack(authenticator(mux)).ServeHTTP(w, r)
		}
	})

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", c.Server.Port),
		Handler:      finalHandler,
		ReadTimeout:  c.Server.TimeoutRead,
		WriteTimeout: c.Server.TimeoutWrite,
		IdleTimeout:  c.Server.TimeoutIdle,
	}

	closed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		l.Info().Msgf("Shutting down server %v", s.Addr)

		ctx, cancel := context.WithTimeout(context.Background(), c.Server.TimeoutIdle)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			l.Error().Err(err).Msg("Server shutdown failure")
		}

		close(closed)
	}()

	l.Info().Msgf("Starting server %v", s.Addr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		l.Fatal().Err(err).Msg("Server startup failure")
	}

	<-closed
	l.Info().Msgf("Server shutdown successfully")
}

func openDatabase(path string, l *logger.Logger, gl gormlogger.LogLevel) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gl),
	})
	if err != nil {
		l.Fatal().Err(err).Msg("Failed to connect to the database")
	}
	return db
}

func migrate(db *gorm.DB, l *logger.Logger) {
	if err := resource.AutoMigrateAll(db); err != nil {
		l.Fatal().Err(err).Msg("Failed to migrate the database")
	}
}
