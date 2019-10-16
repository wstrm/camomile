package dht

import (
	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/optmzr/d7024e-dht/store"
	"github.com/rs/zerolog/log"
)

func (dht *DHT) refreshRequestHandler() {
	for {
		index := <-dht.rt.RefreshCh()

		log.Debug().Msgf("Refresh request for bucket: %d", index)

		id := node.NewIDWithPrefix(dht.me.NodeID, index)

		log.Debug().Msgf("Refresh bucket %d using random ID: %v", index, id)

		_, err := dht.iterativeFindNodes(id)
		if err != nil {
			log.Error().Err(err).
				Msgf("Refresh failed for bucket: %d using random ID: %v",
					index, id)
		}
	}
}

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
			log.Error().Err(err).Msgf("Find nodes network call failed for: %v", request.From.Address)
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

		var touch bool
		switch request.Class {
		case network.StoreClassPublish:
			touch = true
		case network.StoreClassReplicate:
			fallthrough
		default:
			touch = false
		}

		key := store.KeyFromValue(request.Value)
		centrality := dht.rt.Centrality(node.ID(key))

		dht.db.AddItem(key, request.Value, centrality, k, touch)
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

func (dht *DHT) replicateRequestHandler() {
	for {
		item := <-dht.db.ReplicateCh()

		log.Debug().Msgf("Replicate request on value: %v", item)

		_, err := dht.iterativeStore(item.Value, network.StoreClassReplicate)
		if err != nil {
			log.Error().Err(err).
				Msgf("Replicate event failed for value: %v", item)
		}
	}
}

func (dht *DHT) republishRequestHandler() {
	for {
		item := <-dht.db.RepublishCh()

		log.Debug().Msgf("Republish request on value: %v", item)

		_, err := dht.iterativeStore(item.Value, network.StoreClassPublish)
		if err != nil {
			log.Error().Err(err).
				Msgf("Republish event failed for value: %v", item)
		}
	}
}
