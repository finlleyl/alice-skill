package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/finlleyl/alice-skill/internal/logger"
	"github.com/finlleyl/alice-skill/internal/models"
	"github.com/finlleyl/alice-skill/internal/store"
	"go.uber.org/zap"
)

type app struct {
	store store.Store
}

func newApp(s store.Store) *app {
	return &app{store: s}
}

func (a *app) webhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		logger.Log.Debug("got request with bad method", zap.String("method", r.Method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	logger.Log.Debug("decoding request")
	var req models.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if req.Request.Type != models.TypeSimpleUtterance {
		logger.Log.Debug("unsupported request type", zap.String("type", req.Request.Type))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	var text string

	switch true {
	case strings.HasPrefix(req.Request.Command, "Отправь"):
		username, message := parseSendCommand(req.Request.Command)

		recepientID, err := a.store.FindRecepient(ctx, username)
		if err != nil {
			logger.Log.Debug("cannot find recepient by username", zap.String("username", username), zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = a.store.SaveMessage(ctx, recepientID, store.Message{
			Sender:  req.Session.User.UserID,
			Time:    time.Now(),
			Payload: message,
		})
		if err != nil {
			logger.Log.Debug("cannot save message", zap.String("recepient", recepientID), zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		text = "Сообщение успешно отправлено"

	case strings.HasPrefix(req.Request.Command, "Прочитай"):
		messageIndex := parseReadCommand(req.Request.Command)

		messages, err := a.store.ListMessages(ctx, req.Session.User.UserID)
		if err != nil {
			logger.Log.Debug("cannot load messages for user", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		text = "Для вас нет новых сообщений."
		if len(messages) < messageIndex {
			text = "Такого сообщения не существует."
		} else {
			messageID := messages[messageIndex].ID
			message, err := a.store.GetMessage(ctx, messageID)
			if err != nil {
				logger.Log.Debug("cannot load message", zap.Int64("id", messageID), zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			text = fmt.Sprintf("Сообщение от %s, отправлено %s: %s", message.Sender, message.Time, message.Payload)
		}

	default:
		messages, err := a.store.ListMessages(ctx, req.Session.User.UserID)
		if err != nil {
			logger.Log.Debug("cannot load messages for user", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		text = "Для вас нет новых сообщений."
		if len(messages) > 0 {
			text = fmt.Sprintf("Для вас %d новых сообщений.", len(messages))
		}

		if req.Session.New {
			tz, err := time.LoadLocation(req.Timezone)
			if err != nil {
				logger.Log.Debug("cannot parse timezone")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			now := time.Now().In(tz)
			hour, minute, _ := now.Clock()

			text = fmt.Sprintf("Точное время %d часов, %d минут. %s", hour, minute, text)
		}
	}

	resp := models.Response{
		Response: models.ResponsePayload{
			Text: text,
		},
		Version: "1.0",
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}
	logger.Log.Debug("sending HTTP 200 response")
}

func gzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		h.ServeHTTP(ow, r)
	}
}
