package store

import "testing"
import "time"

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

	storedTestItem, err := GetItem(trueHash)

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
		err := AddItem(testVal[i%len(testVal)], testNodeID)

		if err != nil {
			b.Errorf("some error in AddItem that is not yet defined")
		}
	}
}

func TestStoredKeysAdd(t *testing.T) {
	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	AddKey(trueHash)

	storedTestKey, err := GetRepubTime(trueHash)

	if err != nil {
		t.Errorf("Did not find timem entry in key DB for: %x", trueHash)
	}

	if storedTestKey.IsZero() {
		t.Errorf("Key in DB has no time associated.")
	}
}

func TestEvictItem(t *testing.T) {
	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	testVal := "q"

	var testNodeID NodeID
	copy(testNodeID[:], "w")

	err := AddItem(testVal, testNodeID)
	if err != nil {
		t.Errorf("some error in AddItem that is not yet defined")
	}

	// GetItem returns error if item is not found, this assures that something is inserted before we remove it.
	_, err = GetItem(trueHash)
	if err != nil {
		t.Errorf("Item is not in the DB")
	}

	evictItem(trueHash)

	// GetItem should error which assures that the item is no longer found in the item DB.
	_, err = GetItem(trueHash)
	if err == nil {
		t.Errorf("Expected error since item should be removed at this step")
	}
}

func TestGetItem(t *testing.T) {
	fakeHash := [32]byte{17, 69, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	testVal := "q"

	var testNodeID NodeID
	copy(testNodeID[:], "w")

	err := AddItem(testVal, testNodeID)
	if err != nil {
		t.Errorf("Some error in AddItem that is not yet defined")
	}

	_, err = GetItem(fakeHash)
	if err == nil {
		t.Errorf("Found item that was never inserted.")
	}

	_, err = GetItem(trueHash)
	if err != nil {
		t.Errorf("No such items stored in item DB.")
	}
}

func TestGetRepubTime(t *testing.T) {
	fakeHash := [32]byte{17, 69, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}
	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	AddKey(trueHash)

	storedTestKey, err := GetRepubTime(trueHash)
	if err != nil {
		t.Errorf("Did not find timem entry in key DB for: %x", trueHash)
	}

	if storedTestKey.IsZero() {
		t.Errorf("Key in DB has no time associated.")
	}

	_, err = GetRepubTime(fakeHash)
	if err == nil {
		t.Errorf("Found entry in key DB for value that was never inserted.")
	}
}

func setTimers(expire int, replicate int, republish int) {
	tExpire = time.Duration(expire) * time.Second
	tReplicate = time.Duration(replicate) * time.Second
	tRepublish = time.Duration(republish) * time.Second
}

func TestItemHandler(t *testing.T) {
	setup()

	// Test expire timeout eviction.
	setTimers(1, 200, 200)

	testVal := "q"

	var testNodeID NodeID
	copy(testNodeID[:], "w")

	err := AddItem(testVal, testNodeID)
	if err != nil {
		t.Errorf("some error in AddItem that is not yet defined")
	}

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	// After 2 seconds, check if the item  has been removed.
	timer := time.NewTimer(time.Second * 5)
	<-timer.C

	_, err = GetItem(trueHash)
	if err == nil {
		t.Errorf("Item is still in DB after 2 seconds, handler not working.")
	}

	// Test republish timeout eviction.
	setTimers(200, 200, 1)

	err = AddItem(testVal, testNodeID)
	if err != nil {
		t.Errorf("some error in AddItem that is not yet defined")
	}

	// After 2 seconds, check if the item  has been removed.
	timer = time.NewTimer(time.Second * 3)
	<-timer.C

	_, err = GetItem(trueHash)
	if err == nil {
		t.Errorf("Item is still in DB after 2 seconds, handler not working.")
	}
}

func TestRepublisher(t *testing.T) {
	setTimers(200, 200, 1)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	AddKey(trueHash)

	republishedKey := <-db.ch
	if republishedKey != trueHash {
		t.Errorf("Key did not get republished.")
	}
}

func TestReplication(t *testing.T) {
	setTimers(200, 1, 200)

	trueHash := [32]byte{174, 79, 167, 92, 82, 249, 190, 142, 129, 67, 178, 149, 52, 212, 158, 150, 67, 136, 83, 10, 170, 233, 83, 34, 158, 194, 62, 241, 14, 168, 19, 103}

	AddKey(trueHash)

	replicatedKey := <-db.ch
	if replicatedKey != trueHash {
		t.Errorf("Key did not get replicated")
	}
}
