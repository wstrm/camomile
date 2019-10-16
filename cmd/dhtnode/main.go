package main

import (
	"flag"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"

	"github.com/optmzr/d7024e-dht/dht"
	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
)

const defaultDHTAddress = ":8118"

func flagSplit(flag string) (string, string) {
	if flag == "" {
		return "", ""
	}

	components := strings.Split(flag, "@")
	nodeID := components[0]
	address := components[1]

	return nodeID, address
}

func setupLogger(debug bool, logFilepath string) zerolog.Logger {
	var console io.Writer
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		// Pretty print logs instead of JSON.
		console = zerolog.ConsoleWriter{Out: os.Stderr,
			TimeFormat: time.Stamp,
			NoColor:    true,
		}
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		// Unix timestamps are quicker.
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		fw, err := os.Create(logFilepath)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to open log file at: %s", logFilepath)
		}

		// Wrap the file writer io.Writer in a non-blocking wrapper.
		console = diode.NewWriter(fw, 1000, 10*time.Millisecond, func(missed int) {
			fmt.Printf("Logger dropped %d messages", missed)
		})
	}

	// Create logger with timestamps and the console.
	logger := zerolog.New(console).With().Timestamp().Logger()
	log.Logger = logger // Set as global logger.

	// Make sure the standard logger also uses zerolog.
	stdlog.SetFlags(0)
	stdlog.SetOutput(logger)

	return logger
}

func main() {
	meFlag := flag.String("me", "", "Defaults to an auto generated ID, IP defaults to localhost")
	otherFlag := flag.String("other", "", "Waits for incoming connections if not supplied")
	debugFlag := flag.Bool("debug", false, "Print debug logs")
	logFilepathFlag := flag.String("log", "/tmp/dhtnode.log", "File to output logs to")
	flag.Parse()

	logger := setupLogger(*debugFlag, *logFilepathFlag)

	address, err := net.ResolveUDPAddr("udp", defaultDHTAddress)
	if err != nil {
		log.Fatal().Err(err).Msgf("Unable to resolve UDP address: %s", defaultDHTAddress)
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
			log.Fatal().Err(err).Msgf("Unable to parse node ID from: %s", otherID)
		}

		nodeAddress, err := net.ResolveUDPAddr("udp", otherAddress)
		if err != nil {
			log.Fatal().Err(err).Msgf("Unable to resolve UDP address: %s", otherAddress)
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
			log.Fatal().Err(err).Msgf("Unable to parse node ID from: %s", meID)
		}

		nodeAddress, err := net.ResolveUDPAddr("udp", meAddress)
		if err != nil {
			log.Fatal().Err(err).Msgf("Unable to resolve UDP address: %s", meAddress)
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
	go httpServe(dht)

	err = nw.Listen()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen")
	}
}
