package src

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

func TestLRUEviction(t *testing.T) {
	size := 10
	cache := expirable.NewLRU[int, string](size, nil, time.Hour)

	for i := 0; i < size+5; i++ {
		cache.Add(i, fmt.Sprintf("val-%d", i))
	}

	if cache.Len() != size {
		t.Errorf("Expected cache size %d, got %d", size, cache.Len())
	}

	// The first 5 items should be evicted
	for i := 0; i < 5; i++ {
		if _, ok := cache.Get(i); ok {
			t.Errorf("Key %d should have been evicted", i)
		}
	}

	// The last 10 items should be present
	for i := 5; i < size+5; i++ {
		if _, ok := cache.Get(i); !ok {
			t.Errorf("Key %d should be present", i)
		}
	}
}

func TestCacheTTL(t *testing.T) {
	ttl := 100 * time.Millisecond
	cache := expirable.NewLRU[int, string](10, nil, ttl)

	cache.Add(1, "val-1")

	if _, ok := cache.Get(1); !ok {
		t.Fatal("Key 1 should be present immediately after Add")
	}

	time.Sleep(ttl + 50*time.Millisecond)

	if _, ok := cache.Get(1); ok {
		t.Error("Key 1 should have expired")
	}
}
