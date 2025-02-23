package main

import (
	"database/sql"
	"github.com/finlleyl/alice-skill/internal/logger"
	"github.com/finlleyl/alice-skill/internal/store/pg"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	parseFlags()
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := logger.Initialize(flagLogLevel); err != nil {
		return err
	}

	conn, err := sql.Open("pgx", flagDatabaseURI)
	if err != nil {
		return err
	}

	appInstance := newApp(pg.NewStore(conn))

	logger.Log.Info("Running server", zap.String("address", flagRunAddr))

	return http.ListenAndServe(flagRunAddr, logger.RequestLogger(gzipMiddleware(appInstance.webhook)))
}
