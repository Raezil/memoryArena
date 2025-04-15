package memoryArena

import "testing"

func TestNewObject(t *testing.T) {
	arena, err := NewMemoryArena[int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 5
	_, err = NewObject(arena, obj)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func TestAppendSlice(t *testing.T) {
	arena, err := NewMemoryArena[[]int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 5
	slice := []int{1, 2, 3}
	_, err = AppendSlice(&obj, arena, &slice)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
func BenchmarkAppendSlice(b *testing.B) {
	// Setup outside the loop
	// Arena needs enough space for the *resulting* slice after append.
	// 100 bytes is likely too small. Increase significantly for benchmark robustness.
	arenaSize := 10000
	arena, err := NewMemoryArena[[]int](arenaSize)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err) // Use b.Fatalf for setup errors
	}
	initialSlice := []int{1, 2, 3}
	objToAppend := 5

	b.ResetTimer() // Start timing after setup

	for i := 0; i < b.N; i++ {
		arena.Reset() // Reset arena at the start of each iteration

		// Create a copy of the initial slice state for this iteration
		currentSlice := make([]int, len(initialSlice))
		copy(currentSlice, initialSlice)
		// Important: Allocate the *initial* slice copy within the reset arena
		// so AppendSlice operates on an arena-managed slice.
		// If we don't do this, the first AppendSlice might try to reallocate
		// the initial slice from outside the arena, which isn't the point.
		slicePtrInArena, err := NewObject(arena, currentSlice)
		if err != nil {
			b.Fatalf("Iteration %d: Failed to allocate initial slice copy in arena: %v", i, err)
		}
		currentSliceInArena := *slicePtrInArena // Get the slice managed by the arena

		// Perform the operation to be benchmarked
		_, err = AppendSlice(&objToAppend, arena, &currentSliceInArena) // Pass pointer to the arena-managed slice
		if err != nil {
			// Use b.Fatalf as error within loop invalidates benchmark run
			b.Fatalf("Iteration %d: AppendSlice failed: %v", i, err)
		}
	}
}

func BenchmarkNewObject(b *testing.B) {
	// Setup outside the loop
	// Adjust size as needed. 100 might be small depending on b.N and overhead.
	// Calculate required size based on b.N if possible, or use a large enough fixed size.
	// For simplicity, let's use a larger fixed size for the benchmark.
	arenaSize := 10000 // Or calculate based on b.N * estimated size per object
	arena, err := NewMemoryArena[int](arenaSize)
	if err != nil {
		b.Fatalf("Error creating arena: %v", err) // Use b.Fatalf for setup errors
	}
	obj := 5

	b.ResetTimer() // Start timing after setup

	for i := 0; i < b.N; i++ {
		arena.Reset() // Reset arena at the start of each iteration
		_, err = NewObject(arena, obj)
		if err != nil {
			// Use b.Fatalf as error within loop invalidates benchmark run
			b.Fatalf("Iteration %d: NewObject failed: %v", i, err)
		}
	}
}

func TestInsertMap(t *testing.T) {
	arena, err := NewMemoryArena[map[string]int](100)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	obj := 4
	slice := map[string]int{"1": 1, "2": 2, "3": 3}
	_, err = InsertMap(&obj, arena, &slice, "4")
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
func BenchmarkInsertMap(b *testing.B) {
	// Setup outside the loop
	arena, err := NewMemoryArena[map[string]int](1000) // Keep original size or adjust if needed per op
	if err != nil {
		b.Fatalf("Error creating arena: %v", err) // Use b.Fatalf for setup errors
	}
	initialMap := map[string]int{"1": 1, "2": 2, "3": 3}
	keyToInsert := "4"
	valueToInsert := 4

	b.ResetTimer() // Start timing after setup

	for i := 0; i < b.N; i++ {
		arena.Reset() // Reset arena at the start of each iteration

		// Create a copy of the initial map for this iteration to avoid modifying the base map
		currentMap := make(map[string]int, len(initialMap))
		for k, v := range initialMap {
			currentMap[k] = v
		}

		// Perform the operation to be benchmarked
		_, err = InsertMap(&valueToInsert, arena, &currentMap, keyToInsert)
		if err != nil {
			// Use b.Fatalf as error within loop invalidates benchmark run
			b.Fatalf("Iteration %d: InsertMap failed: %v", i, err)
		}
	}
}
