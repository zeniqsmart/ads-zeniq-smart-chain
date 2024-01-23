package fuzz

import (
	"testing"
)

// go test -tags cppbtree -c -coverpkg github.com/zeniqsmart/ads-zeniq-smart-chain/indextree .
// RANDFILE=~/Downloads/rocksdb.tgz RANDCOUNT=10000 ./fuzz.test -test.coverprofile a.out

func Test1(t *testing.T) {
	runTest(DefaultConfig)
}
