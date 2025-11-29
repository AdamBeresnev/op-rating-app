package httputil

import (
	"log/slog"
	"net/http"
)

func InternalServerError(w http.ResponseWriter, msg string, err error) {
	slog.Error(msg, "error", err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

func BadRequest(w http.ResponseWriter, msg string, err error) {
	if err != nil {
		slog.Warn("bad request", "message", msg, "error", err)
	} else {
		slog.Warn("bad request", "message", msg)
	}
	http.Error(w, msg, http.StatusBadRequest)
}

func NotFound(w http.ResponseWriter, msg string, err error) {
	if err != nil {
		slog.Warn("not found", "message", msg, "error", err)
	} else {
		slog.Warn("not found", "message", msg)
	}
	http.Error(w, msg, http.StatusNotFound)
}
