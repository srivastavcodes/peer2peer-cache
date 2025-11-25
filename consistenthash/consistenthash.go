package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash represents the hashing algorithm used.
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int
	keys     []int // sorted
	hashMap  map[int]string
}

// TODO: add functional opts

func New(replicas int, hashFn Hash) *Map {
	mp := &Map{
		hash:     hashFn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if mp.hash == nil {
		mp.hash = crc32.ChecksumIEEE
	}
	return mp
}

// Add adds provided keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get return the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if m.IsEmpty() {
		return ""
	}
	var (
		hash = int(m.hash([]byte(key)))
		idx  = sort.Search(len(m.keys), func(i int) bool {
			return m.keys[i] >= hash
		})
	)
	// means we cycled back to the first replica
	if idx == len(m.keys) {
		idx = 0
	}
	return m.hashMap[m.keys[idx]]
}

// IsEmpty returns true if there are no items available.
func (m *Map) IsEmpty() bool {
	return len(m.keys) == 0
}
