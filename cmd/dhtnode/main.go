package main

import (
	"flag"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"

	"github.com/optmzr/d7024e-dht/ctl"
	"github.com/optmzr/d7024e-dht/dht"
	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
)

const defaultAddress = ":8118"

func rpcServe(dht *dht.DHT) {
	api := ctl.NewAPI(dht)

	err := rpc.Register(api)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to register RPC API")
	}

	rpc.HandleHTTP()
	addr := ":1234"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to listen on: %s", addr)
	}
	log.Info().Msgf("RPC listening on: %s", addr)

	err = http.Serve(l, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to serve at RPC listener")
	}
}

func flagSplit(flag string) (string, string) {
	if flag == "" {
		return "", ""
	}

	components := strings.Split(flag, "@")
	nodeID := components[0]
	address := components[1]

	return nodeID, address
}

func main() {
	meFlag := flag.String("me", "", "Defaults to an auto generated ID, IP defaults to localhost")
	otherFlag := flag.String("other", "", "Waits for incoming connections if not supplied")
	debugFlag := flag.Bool("debug", false, "Print debug logs")
	logFilepathFlag := flag.String("log", "/tmp/dhtnode.log", "File to output logs to")
	flag.Parse()

	var console io.Writer
	if *debugFlag {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		// Pretty print logs instead of JSON.
		console = zerolog.ConsoleWriter{Out: os.Stderr,
			TimeFormat: time.Stamp,
			NoColor:    true,
		}
	} else {
		logFilepath := *logFilepathFlag

		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		// Unix timestamps are quicker.
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		fw, err := os.Create(logFilepath)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Failed to open log file at: %s", logFilepath)
		}
		console = fw
	}

	// Wrap the console io.Writer in a non-blocking wrapper.
	writer := diode.NewWriter(console, 1000, 10*time.Millisecond, func(missed int) {
		fmt.Printf("Logger dropped %d messages", missed)
	})

	// Create logger with timestamps and the non-blocking writer.
	logger := zerolog.New(writer).With().Timestamp().Logger()
	log.Logger = logger // Set as global logger.

	// Make sure the standard logger also uses zerolog.
	stdlog.SetFlags(0)
	stdlog.SetOutput(logger)

	address, err := net.ResolveUDPAddr("udp", defaultAddress)
	if err != nil {
		log.Fatal().
			Err(err).
			Msgf("Unable to resolve UDP address: %s", defaultAddress)
	}

	var others []route.Contact

	otherID, otherAddress := flagSplit(*otherFlag)
	if (otherID == "") || (otherAddress == "") {
		others = []route.Contact{
			route.Contact{
				NodeID:  node.NewID(),
				Address: *address,
			},
		}
	} else {
		nodeID, err := node.IDFromString(otherID)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Unable to parse node ID from: %s", otherID)
		}

		nodeAddress, err := net.ResolveUDPAddr("udp", otherAddress)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Unable to resolve UDP address: %s", otherAddress)
		}

		others = []route.Contact{
			route.Contact{
				NodeID:  nodeID,
				Address: *nodeAddress,
			},
		}
	}

	var me route.Contact

	meID, meAddress := flagSplit(*meFlag)
	if (meID == "") || (meAddress == "") {
		me = route.Contact{
			NodeID:  node.NewID(),
			Address: *address,
		}
	} else {
		nodeID, err := node.IDFromString(meID)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Unable to parse node ID from: %s", meID)
		}

		nodeAddress, err := net.ResolveUDPAddr("udp", meAddress)
		if err != nil {
			log.Fatal().
				Err(err).
				Msgf("Unable to resolve UDP address: %s", meAddress)
		}

		me = route.Contact{
			NodeID:  nodeID,
			Address: *nodeAddress,
		}
	}

	// Add the short node ID to the logger.
	log.Logger = logger.With().Str("nodeid", me.NodeID.String()[:6]).Logger()

	// Print the whole ID:
	log.Info().Msgf("My ID is: %v", me.NodeID)

	nw, err := network.NewUDPNetwork(me)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize network")
	}

	dht, err := dht.New(me, others, nw)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize DHT")
	}

	go rpcServe(dht)

	err = nw.Listen()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen")
	}
}
