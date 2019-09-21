package store

import "testing"
//import "strconv"
//import "fmt"
//import "golang.org/x/crypto/blake2b"
//import "encoding/hex"



func TestItemsAdd(t *testing.T) {
	testVal := "q"

	var testNodeID NodeID
	copy(testNodeID[:], "w")

	err := AddItem(testVal, testNodeID)
	if err != nil {
		t.Errorf("some error in AddItem that is not yet defined")
	}

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}


	// TODO: Remove locks and implement Get function in store.go.
	db.items.RLock()
	storedTestItem, found := db.items.m[trueHash]
	db.items.RUnlock()
	if !found {
		t.Errorf("did not find entry in hash table for: %x", trueHash)
	}

	if !(testVal == storedTestItem.value) {
		t.Errorf("value of item does not match original value.\nExpected: %x\nGot: %x", testVal, storedTestItem.value)
	}

	if !(testNodeID == storedTestItem.origPub) {
		t.Errorf("NodeID stored does not match original NodeID.\nExpected: %x\nGot: %x", testNodeID, storedTestItem.origPub)
	}

	if storedTestItem.expire.IsZero() || storedTestItem.republish.IsZero(){
		t.Errorf("Expire or republish time is not set")
	}
}

func BenchmarkAddItem(b *testing.B) {

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

	var testNodeID NodeID
	copy(testNodeID[:], "w")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddItem(testVal[i%len(testVal)], testNodeID)
	}
}

func TestStoredKeysAdd(t *testing.T){
	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	AddKey(trueHash)

	// TODO: remove locks and replace with Get function in store.go.
	db.keys.RLock()
	storedTestKey, found := db.keys.m[trueHash]
	db.keys.RUnlock()

	if !found {
		t.Errorf("Did not find entry in key hash table for: %x", trueHash)
	}

	if storedTestKey.IsZero() {
		t.Errorf("Key in DB has no time associated.")
	}
}
