package fuzzhis

import (
	"testing"
)

// go test -c -coverpkg github.com/zeniqsmart/moeingads/indextree .
// RANDFILE=~/Downloads/rocksdb.tgz RANDCOUNT=10000 ./fuzzhis.test -test.coverprofile a.out

func Test1(t *testing.T) {
	runTest(DefaultConfig)
}
