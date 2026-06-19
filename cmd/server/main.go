package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"

	"github.com/Irongoshan-ux/gophkeeper/internal/app"
	"github.com/Irongoshan-ux/gophkeeper/internal/config"
	"github.com/Irongoshan-ux/gophkeeper/internal/repository"
	"github.com/Irongoshan-ux/gophkeeper/internal/service"
	"github.com/Irongoshan-ux/gophkeeper/pkg/version"
)

func runMigrations(dsn, migrationsPath string) error {
	m, err := migrate.New("file://"+filepath.ToSlash(migrationsPath), dsn)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func main() {
	versionFlag := flag.Bool("version", false, "print build info")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Info())
		return
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.LoadServer()
	if err != nil {
		logger.Fatal().Err(err).Msg("load config")
	}
	if cfg.DatabaseDSN == "" {
		logger.Fatal().Msg("DATABASE_DSN is required")
	}

	migrationsPath := "migrations"
	if p := os.Getenv("MIGRATIONS_PATH"); p != "" {
		migrationsPath = p
	}
	if err := runMigrations(cfg.DatabaseDSN, migrationsPath); err != nil {
		logger.Fatal().Err(err).Msg("run migrations")
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		logger.Fatal().Err(err).Msg("connect database")
	}
	defer pool.Close()

	users, secrets := repository.NewPostgresRepositories(pool)
	authSvc := service.NewAuthService(users, cfg.JWTSecret)
	secretSvc := service.NewSecretService(secrets)

	srv, err := app.NewGRPCServer(cfg, authSvc, secretSvc, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("create grpc server")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info().Str("address", cfg.Address).Bool("tls", cfg.EnableTLS).Msg("starting gRPC server")
	if err := app.RunGRPC(ctx, srv, cfg.Address); err != nil {
		logger.Fatal().Err(err).Msg("grpc server stopped")
	}
}
