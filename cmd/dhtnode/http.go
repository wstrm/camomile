package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/optmzr/d7024e-dht/dht"
	cdht "github.com/optmzr/d7024e-dht/dht"
	"github.com/optmzr/d7024e-dht/store"
)

const defaultHTTPAddress = ":8080"

type httpHandler struct {
	http.Handler
	dht *dht.DHT
}

func writeError(w http.ResponseWriter, err error, msg string, code int) {
	log.Error().Err(err).Msg(msg)
	http.Error(w, msg, code)
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet: // Get value from DHT.
		key, err := store.KeyFromString(strings.TrimPrefix(r.URL.Path, "/"))
		if err != nil {
			writeError(w, err, "Cannot decode key as hex",
				http.StatusBadRequest)
			return
		}

		value, sender, err := h.dht.Get(key)
		if err != nil {
			writeError(w, err, "Failed to get value by key in DHT",
				http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Value: %s (from: %v)", value, sender)

	case http.MethodPost: // Save value in DHT.
		value := r.PostFormValue("value")
		if value == "" {
			writeError(w, errors.New("no value in form"),
				"Failed to read value in request",
				http.StatusBadRequest)
			return
		}

		key, err := h.dht.Put(value)
		if err != nil {
			writeError(w, err, "Failed to put value in DHT",
				http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Key: %v", key)

	case http.MethodDelete: // Forget value in DHT.
		w.WriteHeader(http.StatusNotImplemented)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func newHTTPHandler(dht *cdht.DHT) *httpHandler {
	return &httpHandler{dht: dht}
}

func newHTTPMux(dht *cdht.DHT) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", newHTTPHandler(dht))
	return mux
}

func httpServe(dht *cdht.DHT) {
	handler := newHTTPHandler(dht)
	log.Info().Msgf("HTTP listening on: %s", defaultHTTPAddress)
	err := http.ListenAndServe(defaultHTTPAddress, handler)
	if err != nil {
		log.Fatal().Err(err).Msgf("HTTP failed to listen on: %s", defaultHTTPAddress)
	}
}
