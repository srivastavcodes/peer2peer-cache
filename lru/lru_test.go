package lru

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type simpleStruct struct {
	int
	string
}

type complexStruct struct {
	int
	simpleStruct
}

var getTests = []struct {
	name     string
	keyToGet any
	keyToAdd any
	expected bool
}{
	{"string_hit", "myKey", "myKey", true},
	{"string_miss", "missed", "myKey", false},
	{"simple_struct_hit",
		simpleStruct{1, "two"},
		simpleStruct{1, "two"},
		true,
	},
	{"simple_struct_miss",
		simpleStruct{0, "missed_again"},
		simpleStruct{1, "two"},
		false,
	},
	{"complex_struct_hit",
		complexStruct{1, simpleStruct{2, "four"}},
		complexStruct{1, simpleStruct{2, "four"}},
		true,
	},
}

func TestGet(t *testing.T) {
	for _, tt := range getTests {
		lru := NewLru(0)

		lru.Add(tt.keyToAdd, 1234)
		val, ok := lru.Get(tt.keyToGet)
		require.Equalf(t, tt.expected, ok, "TestGet: %s: expected cache hit but missed", tt.name)
		if tt.expected {
			require.Equalf(t, 1234, val, "TestGet: %s: value mismatch; want=%v, got=%v", tt.name, 1234, val)
		}
	}
}

func TestRemove(t *testing.T) {
	lru := NewLru(0)

	lru.Add("myKey", 1234)
	val, ok := lru.Get("myKey")
	require.True(t, ok, "TestRemove: Get didn't return any match")
	require.Equalf(t, 1234, val, "TestRemove: Get didn't return expected value; want=%v, got=%v", 1234, val)

	lru.Remove("myKey")
	_, ok = lru.Get("myKey")
	require.False(t, ok, "TestRemove: Get returned a removed value")
}

func TestEvict(t *testing.T) {
	evictedKey := make([]Key, 0)
	onEvictedFunc := func(key Key, value any) {
		evictedKey = append(evictedKey, key)
	}
	lru := NewLru(20)

	lru.OnEvicted = onEvictedFunc
	for i := 0; i < 25; i++ {
		lru.Add(fmt.Sprintf("key_%d", i), i*2)
	}
	require.Equalf(t, 5, len(evictedKey), "got=%d, want=%d", len(evictedKey), 5)

	require.Equalf(t, evictedKey[0], "key_0", "got=%d, want=key_0", evictedKey[0])
	require.Equalf(t, evictedKey[1], "key_1", "got=%d, want=key_1", evictedKey[1])
	require.Equalf(t, evictedKey[2], "key_2", "got=%d, want=key_2", evictedKey[2])
	require.Equalf(t, evictedKey[3], "key_3", "got=%d, want=key_3", evictedKey[3])
	require.Equalf(t, evictedKey[4], "key_4", "got=%d, want=key_4", evictedKey[4])
}
