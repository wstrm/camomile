package store

import (
	"encoding/hex"
	"fmt"

	"github.com/optmzr/d7024e-dht/node"
)

type Key node.ID

func KeyFromString(str string) (key Key, err error) {
	h, err := hex.DecodeString(str)
	if err != nil {
		err = fmt.Errorf("cannot decode hex string as key: %w", err)
		return
	}

	copy(key[:], h)
	return
}

func (k Key) String() string {
	return hex.EncodeToString(k[:])
}
