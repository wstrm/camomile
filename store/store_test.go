package store

import "testing"
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
}

/*func testStoredKeysAdd(){

}*/
