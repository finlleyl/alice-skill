package logger

import (
	"go.uber.org/zap"
	"net/http"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}

func RequestLogger(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Log.Debug("got incoming request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		h(w, r)
	})
}
