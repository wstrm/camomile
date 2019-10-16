package store

import (
	"bytes"
	"testing"
	"time"

	"github.com/optmzr/d7024e-dht/node"
)

func TestItemsAdd(t *testing.T) {
	iHTicker := time.NewTicker(time.Second)
	rHTicker := time.NewTicker(time.Second)
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400, iHTicker, rHTicker)

	testVal := "q"
	testKey := KeyFromValue(testVal)

	db.AddItem(testKey, testVal, 1, 1, true)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	storedTestItem, err := db.GetItem(trueHash)

	if err != nil {
		t.Errorf("did not find entry in hash table for: %x", trueHash)
	}

	if !(testVal == storedTestItem.Value) {
		t.Errorf("value of item does not match original value.\nExpected: %x\nGot: %x", testVal, storedTestItem.Value)
	}

	// No touchy.
	testItemCopy := storedTestItem
	db.AddItem(testKey, "something else", 1, 1, false)
	storedTestItem, _ = db.GetItem(trueHash)

	if testItemCopy.Value != storedTestItem.Value {
		t.Errorf("unexpected value, should be the same")
	}
}

func BenchmarkAddItem(b *testing.B) {
	iHTicker := time.NewTicker(time.Second)
	rHTicker := time.NewTicker(time.Second)
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400, iHTicker, rHTicker)

	testVal := []string{
		"fearlessness",
		"purification of one's existence",
		"cultivation of spritual knowledge",
		"charity",
		"self-control",
		"performance of sacrifice",
		"study of the Vedas",
		"austerity and simplicity",
		"non-violence",
		"truthfulness",
		"freedom from anger",
		"renunciation",
		"tranquility",
		"aversion to faultfinding",
		"compassion and freedom from covetousness",
		"gentleness",
		"modesty and steady determination",
		"vigor",
		"forgiveness",
		"fortitude",
		"cleanliness",
		"freedom from envy and the passion for honor",
	}

	var testKey []Key
	for _, val := range testVal {
		testKey = append(testKey, KeyFromValue(val))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.AddItem(testKey[i%len(testKey)], testVal[i%len(testVal)], 1, 1, false)
	}
}

func TestTruncate(t *testing.T) {
	tooLongString := "9Tf2YFM1NLOxCVWg3e5lclDBPqEV0yzQGHhc41ZUoWTy9maE5hzPyWBgmwMWhg1yM1hb572ZXdXEGjoQvyNT8exx6fikCiFmJQcPBdCcw9rzlR4BseKtyixbeRhh9NF0AWoltgMVPJdYPSgWHYEUlPAdYFCAvlRs5Vumziu2niuPWzhTfzy9RDAfB1Tqt6mHPu9Cxsq1oSZUxltamshva8N2qoc4Rt5qoOoVxMyRxq21WcJ7xXVTHmd1EzpyJ31bnvoiN8zdtc0zPKQ3ddNkuCnRoJzQ78FqPSsXM6DgpNeMcaGFpPwj65hLa2gga4L8N7POF7rZdJJY8vyKIc8b6fLVlrMBlAHuIrrVzjhYw1tuGr26p1TIiV6jfYHPZkZiF5vQCeuN95uCDuP7uJOQUlo4J19pUw2sNB18mMCA7XFYnH4Ys1esF4ordeWkaJ6jLlS3ZThFsfVAVhRzke70ZQUWsWJD6LPJQjILZoffj3hpxlw7FlOeTqpPeHvAyZXX6MTNv95hbU0dWDa6vaUrO3ICVyTHsAr46CpvQMA8kbnfU6szKe1kTgJHvSmL8N9sqcPzd4eMaBtfGUoMBZgHpx18NeaAmx3sZ8RM1gMLDMCO5R0CeW8EsiLkoal4W1bG2nOECi4sGzX22LWcEU1QeuQbn5uFj8oVA8qmCN1cBQreo5cx0AXT0oSMnnuvelJBavHMU8CUjsawq7mUDuzm0M9dBYnXb2INbctkduN5jzAmo1F4ZqAZBOUH2FIr9A8U7bBShtlynWiV8PXepDMXN22kCZ2MRZ7CDkbV4OdFey6MZvbXx9LHZQ8Q4EjQ4FGjV1S0vbrThMVHRNzrjcwWvvZMCDSjE5Ct5d08nJKQ7vZVSdAihVNCyXFVxQIXr8AeFMk6cJDS4E3fbOo9YKJrRWawxJ2h4Q87dLqszVyAo1yJSQawTtinRdq1pogY578J8iMbegqqgLYABrxxnEVU0J2prsx4kGkpMaQRtgggusjA1I46CUmVSsPU3vGB"

	maxThousandCharString := truncate(tooLongString)
	if len(maxThousandCharString) > 1000 {
		t.Errorf("Truncate produces a too long string, more than 1000 chars")
	}

	tenCharString := "0123456789"

	sameString := truncate(tenCharString)
	if len(sameString) != len(tenCharString) {
		t.Errorf("Truncate truncates before 1000 chars")
	}
}

func TestStoredKeysAdd(t *testing.T) {
	iHTicker := time.NewTicker(time.Second)
	rHTicker := time.NewTicker(time.Second)
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400, iHTicker, rHTicker)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	testVal := "q"

	db.AddLocalItem(trueHash, testVal)

	storedLocalItem, ok := getLocalItem(db, trueHash)
	if !ok {
		t.Errorf("Did not find item entry in key DB for: %x", trueHash)
	}

	if storedLocalItem.republish.IsZero() {
		t.Errorf("Key in DB has no time associated.")
	}
}

func TestEvictItem(t *testing.T) {
	iHTicker := time.NewTicker(time.Second)
	rHTicker := time.NewTicker(time.Second)
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400, iHTicker, rHTicker)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	testVal := "q"

	var testNodeID node.ID
	copy(testNodeID[:], "w")

	db.AddItem(KeyFromValue(testVal), testVal, 2, 1, false)

	// GetItem returns error if item is not found, this assures that something is inserted before we remove it.
	_, err := db.GetItem(trueHash)
	if err != nil {
		t.Errorf("Item is not in the DB")
	}

	db.evictRemoteItem(trueHash)

	// GetItem should error which assures that the item is no longer found in the item DB.
	_, err = db.GetItem(trueHash)
	if err == nil {
		t.Errorf("Expected error since item should be removed at this step")
	}
}

func TestGetItem(t *testing.T) {
	iHTicker := time.NewTicker(time.Second)
	rHTicker := time.NewTicker(time.Second)
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400, iHTicker, rHTicker)

	fakeHash := [32]byte{17, 69, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	testVal := "q"

	var testNodeID node.ID
	copy(testNodeID[:], "w")

	db.AddItem(KeyFromValue(testVal), testVal, 1, 1, false)

	_, err := db.GetItem(fakeHash)
	if err == nil {
		t.Errorf("Found item that was never inserted.")
	}

	_, err = db.GetItem(trueHash)
	if err != nil {
		t.Errorf("No such items stored in item DB.")
	}
}

func getLocalItem(db *Database, key Key) (localItem, bool) {
	db.localItems.RLock()
	defer db.localItems.RUnlock()
	item, ok := db.localItems.m[key]
	return item, ok
}

func TestGetRepubTime(t *testing.T) {
	iHTicker := time.NewTicker(time.Second)
	rHTicker := time.NewTicker(time.Second)
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400, iHTicker, rHTicker)

	fakeHash := [32]byte{17, 69, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	testVal := "q"

	db.AddLocalItem(trueHash, testVal)

	storedLocalItem, ok := getLocalItem(db, trueHash)
	if !ok {
		t.Error("key doesn't exist in db")
	}

	if storedLocalItem.republish.IsZero() {
		t.Error("key in DB has no time associated.")
	}

	_, ok = getLocalItem(db, fakeHash)
	if ok {
		t.Error("Found entry in key DB for value that was never inserted.")
	}
}

func TestItemHandler(t *testing.T) {
	rHTicker := time.NewTicker(time.Second)

	tch := make(chan time.Time)
	iHTicker := &time.Ticker{
		C: tch,
	}

	tick := make(chan struct{})
	go func(tch chan time.Time, tick chan struct{}) {
		// Add 1000 hours to mock "now". This will make the item to expire
		// immediately.
		<-tick
		tch <- time.Now().Add(1000 * time.Hour)
	}(tch, tick)

	testVal := "q"
	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	db := NewDatabase(time.Second*86410, time.Second*3600, time.Second*86400, iHTicker, rHTicker)

	db.AddItem(KeyFromValue(testVal), testVal, 1, 1, false)
	_, err := db.GetItem(trueHash)
	if err != nil {
		t.Error("no item was added to db")
	}

	tick <- struct{}{} // Item should have been added, make the ticker tick.

	for start := time.Now(); time.Since(start) < time.Second; {
		_, err := db.GetItem(trueHash)
		if err != nil {
			return // Done, item was removed.
		}
	}

	t.Error("item is still in db")
}

func TestRepublisher(t *testing.T) {
	iHTicker := time.NewTicker(time.Second)

	tch := make(chan time.Time)
	rHTicker := &time.Ticker{
		C: tch,
	}

	tick := make(chan struct{})
	go func(tch chan time.Time, tick chan struct{}) {
		// Add 1000 hours to mock "now". This will make the item to expire
		// immediately.
		<-tick
		tch <- time.Now().Add(1000 * time.Hour)
	}(tch, tick)

	db := NewDatabase(time.Second*86410, time.Second*3600, time.Second*86400, iHTicker, rHTicker)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	testVal := "q"

	db.AddLocalItem(trueHash, testVal)
	tick <- struct{}{} // Item should have been added, make the ticker tick.

	republished := <-db.republishCh
	if republished.Value != testVal {
		t.Errorf("LocalItem did not get republished.")
	}
}

func TestReplication(t *testing.T) {
	iHTicker := time.NewTicker(time.Second)

	tch := make(chan time.Time)
	rHTicker := &time.Ticker{
		C: tch,
	}

	tick := make(chan struct{})
	go func(tch chan time.Time, tick chan struct{}) {
		// Add 1000 hours to mock "now". This will make the item to expire
		// immediately.
		<-tick
		tch <- time.Now().Add(1000 * time.Hour)
	}(tch, tick)

	db := NewDatabase(time.Second*86400, time.Second*0, time.Second*86400, iHTicker, rHTicker)

	testVal := "q"

	db.AddItem(KeyFromValue(testVal), testVal, 1, 1, false)
	tick <- struct{}{} // Item should have been added, make the ticker tick.

	replicated := <-db.replicateCh
	if replicated.Value != testVal {
		t.Errorf("Key did not get replicated")
	}
}

func TestKeyFromString(t *testing.T) {
	validKey := "53f2a6d618d66a05378bc38aee2a17c82b0310d8574200ce684539255416dfe3"
	invalidKey := "ABC, du är mina tankar"

	h, err := KeyFromString(validKey)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expH := Key{83, 242, 166, 214, 24, 214, 106, 5, 55, 139, 195, 138, 238, 42, 23, 200, 43, 3, 16, 216, 87, 66, 0, 206, 104, 69, 57, 37, 84, 22, 223, 227}

	if !bytes.Equal(h[:], expH[:]) {
		t.Errorf("unexpected key, got: %v, expected: %v", h, expH)
	}

	_, err = KeyFromString(invalidKey)
	if err == nil {
		t.Errorf("expected error for string: %s", invalidKey)
	}
}

func TestRepublishCh(t *testing.T) {
	iHTicker := time.NewTicker(time.Second)
	rHTicker := time.NewTicker(time.Second)
	db := NewDatabase(time.Second*86400, time.Second*0, time.Second*86400, iHTicker, rHTicker)

	returnedChan := db.RepublishCh()
	go func() { returnedChan <- Item{} }()
	<-returnedChan
}

func TestReplicateCh(t *testing.T) {
	iHTicker := time.NewTicker(time.Second)
	rHTicker := time.NewTicker(time.Second)
	db := NewDatabase(time.Second*86400, time.Second*0, time.Second*86400, iHTicker, rHTicker)

	returnedChan := db.ReplicateCh()
	go func() { returnedChan <- Item{} }()
	<-returnedChan
}

func TestForgetItem(t *testing.T) {
	iHTicker := time.NewTicker(time.Second)
	rHTicker := time.NewTicker(time.Second)
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400, iHTicker, rHTicker)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	testVal := "q"

	db.AddLocalItem(trueHash, testVal)

	_, ok := getLocalItem(db, trueHash)
	if !ok {
		t.Error("expected localItem to be in db")
	}

	db.ForgetItem(trueHash)

	_, ok = getLocalItem(db, trueHash)
	if ok {
		t.Error("expected localItem is still in db, expected error")
	}
}

func TestItemString(t *testing.T) {
	item := Item{Key: [32]byte{}, Value: "q"}
	str := item.String()
	if str != "0000000000000000000000000000000000000000000000000000000000000000: q" {
		t.Errorf("unexpected string: %s", str)
	}
}
