package store

import (
	"fmt"
	"math"
	"sync"

	"github.com/zeniqsmart/moeingads"
	"github.com/zeniqsmart/moeingads/store/types"
	adstypes "github.com/zeniqsmart/moeingads/types"
)

const (
	EvictTryDist = 16
)

type cacheEntry struct {
	height int64
	value  string
}

// RootStore wraps a cache and a moeingads.
// This cache can hold written data from many blocks, and
// when it reaches size limit, the older data are evicted out.
// This cache is write-through, which means the data written
// to it are also written to MoeingADS. When cache misses, we can
// always fetch data from MoeingADS
// A RWMutex protects MoeingADS and cache, making sure when data
// are written to them, no read operation is permitted.
type RootStore struct {
	mtx sync.RWMutex

	cache [adstypes.ShardCount]map[string]cacheEntry
	// controls which KVs can be cached, when this function is nil, all are cached.
	isCacheableKey func(k []byte) bool
	mads           *moeingads.MoeingADS
	// Current height. Each newly-added cache entry carries this value as an 'age-tag'.
	// When entries are evicted, we try to find the oldest one (minimum height)
	height int64
}

var _ types.RootStoreI = &RootStore{}

func NewRootStore(mads *moeingads.MoeingADS, isCacheableKey func(k []byte) bool) *RootStore {
	result := &RootStore{
		isCacheableKey: isCacheableKey,
		mads:           mads,
		height:         -1,
	}
	for i := range result.cache {
		result.cache[i] = make(map[string]cacheEntry)
	}
	return result
}

// One RootStore can be used by many TrunkStores and RabbitStores. There can be only one writter and many
// readers. RootStore is written in TruckStore's WriteBack, which protect the writing process with the write-lock,
// and is read during the life time of RabbitStores. So RabbitStores acquire read-lock at creation and release
// read-lock when closing.
// Writer won't starve: https://stackoverflow.com/questions/46786900/the-implementions-of-rwmutex-lock-in-golang
func (root *RootStore) RLock() {
	root.mtx.RLock()
}
func (root *RootStore) Lock() {
	root.mtx.Lock()
}
func (root *RootStore) RUnlock() {
	root.mtx.RUnlock()
}
func (root *RootStore) Unlock() {
	root.mtx.Unlock()
}

func (root *RootStore) SetHeight(h int64) {
	root.height = h
}

func (root *RootStore) Get(key []byte) []byte {
	ok := false
	var e cacheEntry
	if root.isCacheableKey != nil && root.isCacheableKey(key) {
		shardID := adstypes.GetShardID(key)
		e, ok = root.cache[shardID][string(key)]
	}
	if ok {
		return []byte(e.value)
	} else {
		return root.get(key)
	}
}

func (root *RootStore) GetAtHeight(key []byte, height uint64) []byte {
	e := root.mads.GetEntryAtHeight(key, height)
	if e == nil {
		return nil
	}
	return e.Value
}

func (root *RootStore) get(key []byte) []byte {
	e := root.mads.GetEntry(key)
	if e == nil {
		return nil
	}
	return e.Value
}

func (root *RootStore) PrepareForUpdate(key []byte) {
	root.mads.PrepareForUpdate(key)
}

func (root *RootStore) PrepareForDeletion(key []byte) {
	root.mads.PrepareForDeletion(key)
}

// When the client code writes RootStore, it must follow 'BeginWrite -> Set&Delete -> EndWrite' sequence,
// or it can use the 'Update' function which ensures such a sequence.
func (root *RootStore) BeginWrite() {
	if root.height < 0 {
		panic(fmt.Sprintf("Height is not initialized: %d", root.height))
	}
	root.mads.BeginWrite(root.height)
}

func (root *RootStore) Set(key, value []byte) {
	root.mads.Set(key, value)
	if root.isCacheableKey != nil && root.isCacheableKey(key) {
		root.addToCache([]byte(key), value)
	}
}

func (root *RootStore) Delete(key []byte) {
	root.mads.Delete(key)
	shardID := adstypes.GetShardID(key)
	delete(root.cache[shardID], string(key))
}

func (root *RootStore) EndWrite() {
	root.mads.EndWrite()
}

func (root *RootStore) Update(updater func(db types.SetDeleter)) {
	root.Lock()
	root.BeginWrite()
	updater(root)
	root.EndWrite()
	root.Unlock()
}

func (root *RootStore) addToCache(key, value []byte) {
	shardID := adstypes.GetShardID(key)
	if len(root.cache[shardID]) > moeingads.RootCacheSizeLimit {
		var delK string
		delHeight := int64(math.MaxInt64)
		dist := 0
		//Now we find the oldest entry within EvictTryDist steps.
		//Please note Golang's map-iteration is randomized.
		for k, e := range root.cache[shardID] {
			if delHeight > e.height {
				// find the key with minimum height
				delHeight = e.height
				delK = k
			}
			dist++
			if dist >= EvictTryDist {
				break
			}
		}
		if len(delK) != 0 {
			delete(root.cache[shardID], delK)
		}
	}
	root.cache[shardID][string(key)] = cacheEntry{
		height: root.height,
		value:  string(value),
	}
}

func (root *RootStore) GetTrunkStore(cacheSize int) interface{} {
	return &TrunkStore{
		cache:     NewCacheStore(cacheSize),
		root:      root,
		isWriting: 0,
	}
}

func (root *RootStore) GetReadOnlyTrunkStore(cacheSize int) interface{} {
	return &TrunkStore{
		cache:      NewCacheStore(cacheSize),
		root:       root,
		isWriting:  0,
		isReadOnly: true,
	}
}

func (root *RootStore) GetRootHash() []byte {
	return root.mads.GetRootHash()
}

func (root *RootStore) Close() {
	root.mads.Close()
	for i := range root.cache {
		root.cache[i] = nil
	}
}

func (root *RootStore) ActiveCount() int {
	return root.mads.ActiveCount()
}
