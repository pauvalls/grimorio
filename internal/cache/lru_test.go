package cache

import (
	"sync"
	"testing"
)

func TestLRUCache_BasicGetPut(t *testing.T) {
	cache := NewLRU[string, int](3)

	cache.Put("a", 1)
	cache.Put("b", 2)

	val, ok := cache.Get("a")
	if !ok {
		t.Fatal("expected key 'a' to exist")
	}
	if val != 1 {
		t.Fatalf("expected value 1, got %d", val)
	}

	val, ok = cache.Get("b")
	if !ok {
		t.Fatal("expected key 'b' to exist")
	}
	if val != 2 {
		t.Fatalf("expected value 2, got %d", val)
	}
}

func TestLRUCache_Miss(t *testing.T) {
	cache := NewLRU[string, int](2)

	_, ok := cache.Get("nonexistent")
	if ok {
		t.Fatal("expected key 'nonexistent' to not exist")
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRU[string, int](2)

	cache.Put("a", 1)
	cache.Put("b", 2)
	cache.Put("c", 3) // Should evict "a" (LRU)

	_, ok := cache.Get("a")
	if ok {
		t.Fatal("expected key 'a' to be evicted")
	}

	val, ok := cache.Get("b")
	if !ok {
		t.Fatal("expected key 'b' to exist")
	}
	if val != 2 {
		t.Fatalf("expected value 2, got %d", val)
	}

	val, ok = cache.Get("c")
	if !ok {
		t.Fatal("expected key 'c' to exist")
	}
	if val != 3 {
		t.Fatalf("expected value 3, got %d", val)
	}
}

func TestLRUCache_UpdateExisting(t *testing.T) {
	cache := NewLRU[string, int](2)

	cache.Put("a", 1)
	cache.Put("a", 10)

	val, ok := cache.Get("a")
	if !ok {
		t.Fatal("expected key 'a' to exist")
	}
	if val != 10 {
		t.Fatalf("expected updated value 10, got %d", val)
	}
}

func TestLRUCache_GetPromotesToFront(t *testing.T) {
	cache := NewLRU[string, int](2)

	cache.Put("a", 1)
	cache.Put("b", 2)

	// Access "a" to make it recently used
	cache.Get("a")

	// Now add "c" — "b" should be evicted (LRU) since "a" was just accessed
	cache.Put("c", 3)

	val, ok := cache.Get("a")
	if !ok {
		t.Fatal("expected key 'a' to exist after promotion")
	}
	if val != 1 {
		t.Fatalf("expected value 1, got %d", val)
	}

	_, ok = cache.Get("b")
	if ok {
		t.Fatal("expected key 'b' to be evicted")
	}
}

func TestLRUCache_Remove(t *testing.T) {
	cache := NewLRU[string, int](2)

	cache.Put("a", 1)
	cache.Put("b", 2)

	cache.Remove("a")

	_, ok := cache.Get("a")
	if ok {
		t.Fatal("expected key 'a' to be removed")
	}

	val, ok := cache.Get("b")
	if !ok {
		t.Fatal("expected key 'b' to still exist")
	}
	if val != 2 {
		t.Fatalf("expected value 2, got %d", val)
	}
}

func TestLRUCache_ThreadSafety(t *testing.T) {
	cache := NewLRU[int, int](100)
	var wg sync.WaitGroup
	numGoroutines := 50
	numOperations := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := (id*numOperations + j) % 200
				cache.Put(key, key*2)
				cache.Get(key)
			}
		}(i)
	}

	wg.Wait()

	// Verify cache contains expected entries
	for i := 0; i < 100; i++ {
		val, ok := cache.Get(i)
		if !ok {
			// Some entries may have been evicted — that's OK
			continue
		}
		if val != i*2 {
			t.Fatalf("expected value %d, got %d", i*2, val)
		}
	}
}

func TestLRUCache_ZeroCapacity(t *testing.T) {
	cache := NewLRU[string, int](0)

	cache.Put("a", 1)

	_, ok := cache.Get("a")
	if ok {
		t.Fatal("expected key 'a' to not exist in zero-capacity cache")
	}
}

func TestLRUCache_CapacityOne(t *testing.T) {
	cache := NewLRU[string, int](1)

	cache.Put("a", 1)
	cache.Put("b", 2)

	_, ok := cache.Get("a")
	if ok {
		t.Fatal("expected key 'a' to be evicted")
	}

	val, ok := cache.Get("b")
	if !ok {
		t.Fatal("expected key 'b' to exist")
	}
	if val != 2 {
		t.Fatalf("expected value 2, got %d", val)
	}
}
