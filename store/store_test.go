package store

import "testing"


func testBlake2b(){
	// After testItemsAdd has ran, the stored hash should be the 
	// value of 36ff942f90c129cfe207dc3063f4cd19aaeda211. 
	// This is the hex representation of "q" hashed with blake2b-160.

	if !testItemsMap.
}

func populateTestItem() {
	testItem := item

	// "q" translates to 01110001 in binary.
	v := "q"
	copy(testItem.value[:], v)

	testItem.expire = time.Now().Add(time.Second * 2)
	testItem.republish = time.Now().Add(time.Second * 2)

	// "w" translates to 01110111 in binary.
	s := "w"
	copy(testItem.origPub[:], s)
}

func testItemsAdd() {
	testVal := [255]byte
	copy(testVal[:], "q")

	testNodeID := NodeID
	copy(testNodeID[:], "w")

	testItemsMap := items
	testItemsMap.add(testVal, testNodeID)

	if !testItemsMap[// TODO: extract first key] == the true value of q hashed with blake2b-160
}
