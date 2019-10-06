package store

import (
	"bytes"
	"testing"
	"time"

	"github.com/optmzr/d7024e-dht/node"
)

func TestItemsAdd(t *testing.T) {
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

	testVal := "q"

	var testNodeID node.ID
	copy(testNodeID[:], "w")

	db.AddItem(testVal, testNodeID)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	storedTestItem, err := db.GetItem(trueHash)

	if err != nil {
		t.Errorf("did not find entry in hash table for: %x", trueHash)
	}

	if !(testVal == storedTestItem.Value) {
		t.Errorf("value of item does not match original value.\nExpected: %x\nGot: %x", testVal, storedTestItem.Value)
	}

	if !(testNodeID == storedTestItem.origPub) {
		t.Errorf("NodeID stored does not match original NodeID.\nExpected: %x\nGot: %x", testNodeID, storedTestItem.origPub)
	}

	if storedTestItem.expire.IsZero() || storedTestItem.republish.IsZero() {
		t.Errorf("Expire or republish time is not set")
	}
}

func BenchmarkAddItem(b *testing.B) {
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

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

	var testNodeID node.ID
	copy(testNodeID[:], "w")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.AddItem(testVal[i%len(testVal)], testNodeID)
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
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	testVal := "q"

	db.AddLocalItem(trueHash, testVal)

	storedTestKey, err := db.GetLocalItem(trueHash)
	if err != nil {
		t.Errorf("Did not find timem entry in key DB for: %x", trueHash)
	}

	if storedTestKey.repubTime.IsZero() {
		t.Errorf("Key in DB has no time associated.")
	}
}

func TestEvictItem(t *testing.T) {
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	testVal := "q"

	var testNodeID node.ID
	copy(testNodeID[:], "w")

	db.AddItem(testVal, testNodeID)

	// GetItem returns error if item is not found, this assures that something is inserted before we remove it.
	_, err := db.GetItem(trueHash)
	if err != nil {
		t.Errorf("Item is not in the DB")
	}

	db.evictItem(trueHash)

	// GetItem should error which assures that the item is no longer found in the item DB.
	_, err = db.GetItem(trueHash)
	if err == nil {
		t.Errorf("Expected error since item should be removed at this step")
	}
}

func TestGetItem(t *testing.T) {
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

	fakeHash := [32]byte{17, 69, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	testVal := "q"

	var testNodeID node.ID
	copy(testNodeID[:], "w")

	db.AddItem(testVal, testNodeID)

	_, err := db.GetItem(fakeHash)
	if err == nil {
		t.Errorf("Found item that was never inserted.")
	}

	_, err = db.GetItem(trueHash)
	if err != nil {
		t.Errorf("No such items stored in item DB.")
	}
}

func TestGetRepubTime(t *testing.T) {
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

	fakeHash := [32]byte{17, 69, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	testVal := "q"

	db.AddLocalItem(trueHash, testVal)

	storedTestKey, err := db.GetLocalItem(trueHash)
	if err != nil {
		t.Errorf("Did not find timem entry in key DB for: %x", trueHash)
	}

	if storedTestKey.repubTime.IsZero() {
		t.Errorf("Key in DB has no time associated.")
	}

	_, err = db.GetLocalItem(fakeHash)
	if err == nil {
		t.Errorf("Found entry in key DB for value that was never inserted.")
	}
}

func TestItemHandler(t *testing.T) {
	db := NewDatabase(time.Second*0, time.Second*3600, time.Second*86400)

	testVal := "q"

	var testNodeID node.ID
	copy(testNodeID[:], "w")

	db.AddItem(testVal, testNodeID)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	// After 2 seconds, check if the item  has been removed.
	timer := time.NewTimer(time.Second * 3)
	<-timer.C

	_, err := db.GetItem(trueHash)

	if err == nil {
		t.Errorf("Item is still in DB after 2 seconds, handler not working.")
	}
}

func TestItemHandlerRepub(t *testing.T) {
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*0)
	testVal := "q"

	var testNodeID node.ID
	copy(testNodeID[:], "w")
	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	db.AddItem(testVal, testNodeID)

	// After 2 seconds, check if the item  has been removed.
	timer := time.NewTimer(time.Second * 3)
	<-timer.C

	_, err := db.GetItem(trueHash)
	if err == nil {
		t.Errorf("Item is still in DB after 2 seconds, handler not working.")
	}
}

func TestRepublisher(t *testing.T) {
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*0)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	testVal := "q"

	db.AddLocalItem(trueHash, testVal)

	republishedValue := <-db.ch
	if republishedValue != testVal {
		t.Errorf("LocalItem did not get republished.")
	}
}

func TestReplication(t *testing.T) {
	db := NewDatabase(time.Second*86400, time.Second*0, time.Second*86400)

	testVal := "q"
	var testNodeID node.ID
	copy(testNodeID[:], "w")

	db.AddItem(testVal, testNodeID)

	replicatedValue := <-db.ch
	if replicatedValue != testVal {
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

func TestLocalItemCh(t *testing.T) {
	db := NewDatabase(time.Second*86400, time.Second*0, time.Second*86400)

	returnedChan := db.LocalItemCh()
	go func() { returnedChan <- "abc" }()
	<-returnedChan
}

func TestForgetItem(t *testing.T) {
	db := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	testVal := "q"

	db.AddLocalItem(trueHash, testVal)

	_, err := db.GetLocalItem(trueHash)
	if err != nil {
		t.Errorf("expected localItem to be in db")
	}

	db.ForgetItem(trueHash)

	_, err = db.GetLocalItem(trueHash)
	if err == nil {
		t.Errorf("expected localItem is still in db, expected error")
	}
}
