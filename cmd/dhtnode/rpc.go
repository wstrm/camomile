package main

import (
	"net/http"
	"net/rpc"

	"github.com/optmzr/d7024e-dht/ctl"
	"github.com/optmzr/d7024e-dht/dht"
	"github.com/rs/zerolog/log"
)

const defaultRPCAddress = ":1234"

func rpcServe(dht *dht.DHT) {
	api := ctl.NewAPI(dht)

	err := rpc.Register(api)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register RPC API")
	}
	rpc.HandleHTTP()

	log.Info().Msgf("RPC listening on: %s", defaultRPCAddress)
	err = http.ListenAndServe(defaultRPCAddress, nil)
	if err != nil {
		log.Fatal().Err(err).Msgf("RPC failed to listen on: %s", defaultRPCAddress)
	}
}
