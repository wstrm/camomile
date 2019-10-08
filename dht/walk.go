package dht

import (
	"fmt"

	"github.com/optmzr/d7024e-dht/network"
	"github.com/optmzr/d7024e-dht/node"
	"github.com/optmzr/d7024e-dht/route"
	"github.com/rs/zerolog/log"
)

type awaitChannel struct {
	ch     chan network.Result
	callee route.Contact
}

type awaitResult struct {
	result network.Result
	callee route.Contact
}

func (dht *DHT) walk(call Call) ([]route.Contact, error) {
	nw := dht.nw
	me := dht.me
	target := call.Target()

	// The first α contacts selected are used to create a *shortlist* for the
	// search.
	sl := dht.rt.NClosest(target, α)

	// Keep a map of contacts that has been sent to, to make sure we do not
	// contact the same node multiple times.
	sent := make(map[node.ID]bool)

	// If a cycle results in an unchanged `closest` node, then a FindNode
	// network call should be made to each of the closest nodes that has not
	// already been queried.
	rest := false

	// Contacts holds a sorted (slice) copy of the shortlist.
	contacts := sl.SortedContacts()

	if len(contacts) == 0 {
		// No candidates found in the routing table.
		return contacts, fmt.Errorf("empty routing table")
	}

	// Closest is the node that closest in distance to the target node ID.
	closest := contacts[0]

	for {
		// Holds a slice of channels that are awaiting a response from the
		// network.
		await := []awaitChannel{}

		for i, contact := range contacts {
			if i >= α && !rest {
				break // Limit to α contacts per shortlist.
			}
			if sent[contact.NodeID] || contact.NodeID.Equal(me.NodeID) {
				continue // Ignore already contacted contacts or local node.
			}

			ch, err := call.Do(nw, contact.Address)
			if err != nil {
				log.Error().Err(err).
					Msgf("Unable to dial: %v, removing from candidates...",
						contact.NodeID)

				sl.Remove(contact)
			} else {
				// Mark as contacted.
				sent[contact.NodeID] = true

				// Add to await channel queue.
				await = append(await, awaitChannel{ch: ch, callee: contact})
			}
		}

		results := make(chan awaitResult)
		for _, ac := range await {
			go func(ac awaitChannel) {
				// Redirect all responses to the results channel.
				r := <-ac.ch
				results <- awaitResult{result: r, callee: ac.callee}
			}(ac)
		}

		// Iterate through every result from the responding nodes and add their
		// closest contacts to the shortlist.
		for i := 0; i < len(await); i++ {
			ac := <-results
			result := ac.result
			callee := ac.callee

			if result != nil {
				// Add node so it is moved to the top of its bucket in the
				// routing table.
				go dht.addNode(callee)

				// Add the responding node's closest contacts.
				sl.Add(result.Closest()...)

				// Update callee with intermediate results.
				stop := call.Result(result, callee)
				if stop {
					break // Callee requested that the walk must be stopped.
				}
			} else {
				// Network response timed out.
				log.Warn().
					Msgf("Network response from: %v timed out, removing from candidates...",
						callee.NodeID)

				// Remove the callee from the candidates.
				sl.Remove(callee)
			}
		}

		contacts = sl.SortedContacts()

		if len(contacts) == 0 {
			// No candidates responded and all of them was therefore removed
			// from the shortlist.
			return contacts, fmt.Errorf("no candidates responded")
		}

		first := contacts[0]
		if closest.NodeID.Equal(first.NodeID) {
			// Unchanged closest node from last run, re-run but check all the
			// nodes in the shortlist (and not only the α closest).
			if !rest {
				rest = true
				continue
			}

			// Done. Return the contacts in the shortlist sorted by distance.
			return contacts, nil

		} else {
			// New closest node found, continue iteration.
			closest = first
		}
	}
}
