package store

import (
	"testing"
	"time"

	"github.com/optmzr/d7024e-dht/node"
)

func TestItemsAdd(t *testing.T) {
	db, _ := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

	testVal := "q"

	var testNodeID node.ID
	copy(testNodeID[:], "w")

	db.AddItem(testVal, testNodeID)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	storedTestItem, err := db.GetItem(trueHash)

	if err != nil {
		t.Errorf("did not find entry in hash table for: %x", trueHash)
	}

	if !(testVal == storedTestItem.value) {
		t.Errorf("value of item does not match original value.\nExpected: %x\nGot: %x", testVal, storedTestItem.value)
	}

	if !(testNodeID == storedTestItem.origPub) {
		t.Errorf("NodeID stored does not match original NodeID.\nExpected: %x\nGot: %x", testNodeID, storedTestItem.origPub)
	}

	if storedTestItem.expire.IsZero() || storedTestItem.republish.IsZero() {
		t.Errorf("Expire or republish time is not set")
	}
}

func BenchmarkAddItem(b *testing.B) {
	db, _ := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

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

func TestStoredKeysAdd(t *testing.T) {
	db, _ := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

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
	db, _ := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

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
	db, _ := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

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
	db, _ := NewDatabase(time.Second*86400, time.Second*3600, time.Second*86400)

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
	db, _ := NewDatabase(time.Second*0, time.Second*3600, time.Second*86400)

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
	db, _ := NewDatabase(time.Second*86400, time.Second*3600, time.Second*0)
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
	db, _ := NewDatabase(time.Second*86400, time.Second*3600, time.Second*0)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	testVal := "q"

	db.AddLocalItem(trueHash, testVal)

	republishedItem := <-db.ch
	if republishedItem.value != testVal {
		t.Errorf("LocalItem did not get republished.")
	}
}

func TestReplication(t *testing.T) {
	db, _ := NewDatabase(time.Second*86400, time.Second*0, time.Second*86400)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	testVal := "q"

	db.AddLocalItem(trueHash, testVal)

	replicatedItem := <-db.ch
	if replicatedItem.value != testVal {
		t.Errorf("Key did not get replicated")
	}
}
