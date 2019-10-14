package dht

import (
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/rs/zerolog/log"
)

func (dht *DHT) findValueRequestHandler() {
	for {
		request := <-dht.nw.FindValueRequestCh()

		log.Debug().Msgf("Find value request from: %v", request.From.NodeID)

		// Add node so it is moved to the top of its bucket in the routing
		// table.
		go dht.addNode(request.From)

		var closest []route.Contact
		target := node.ID(request.Key)

		// Try to fetch the value from the local storage.
		item, err := dht.db.GetItem(request.Key)
		if err != nil {
			// No luck.
			// Fetch this nodes contacts that are closest to the requested key.
			closest = dht.rt.NClosest(target, k).SortedContacts()
		} else {
			log.Debug().Msgf("Found value: %s", item.Value)
		}

		err = dht.nw.SendValue(request.Key, item.Value, closest,
			request.SessionID, request.From.Address)
		if err != nil {
			log.Error().Err(err).
				Msgf("Send value network call failed for: %v",
					request.From.Address)
		}
	}
}

func (dht *DHT) findNodesRequestHandler() {
	for {
		request := <-dht.nw.FindNodesRequestCh()

		log.Debug().Msgf("Find node request from: %v", request.From.NodeID)

		// Add node so it is moved to the top of its bucket in the routing
		// table.
		go dht.addNode(request.From)

		// Fetch this nodes contacts that are closest to the requested target.
		closest := dht.rt.NClosest(request.Target, k).SortedContacts()

		err := dht.nw.SendNodes(closest, request.SessionID, request.From.Address)
		if err != nil {
			log.
				Error().
				Err(err).
				Msgf("Find nodes network call failed for: %v",
					request.From.Address)
		}
	}
}

func (dht *DHT) storeRequestHandler() {
	for {
		request := <-dht.nw.StoreRequestCh()

		log.Debug().Msgf("Store value request from: %v", request.From.NodeID)

		// Add node so it is moved to the top of its bucket in the routing
		// table.
		go dht.addNode(request.From)

		dht.db.AddItem(request.Value, request.From.NodeID)
	}
}

func (dht *DHT) pongRequestHandler() {
	for {
		request := <-dht.nw.PongRequestCh()

		log.Printf("Pong request from: %v (%x)",
			request.From.NodeID, request.Challenge)

		// Add node so it is moved to the top of its bucket in the routing
		// table.
		go dht.addNode(request.From)

		err := dht.nw.Pong(
			request.Challenge,
			request.SessionID,
			request.From.Address)
		if err != nil {
			log.Error().Err(err).
				Msgf("Pong network call failed for: %v",
					request.From.Address)
		}
	}
}

func (dht *DHT) republishRequestHandler() {
	for {
		value := <-dht.db.ItemCh()

		log.Debug().Msgf("Republish request on value: %v", value)

		_, err := dht.iterativeStore(value)
		if err != nil {
			log.Error().Err(err).
				Msgf("Republish event failed for value: %v", value)
		}
	}
}
