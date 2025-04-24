<p align="center">
  <img src="https://github.com/user-attachments/assets/2930ba29-f815-492f-ae98-fe0151a2ae12">
</p>

# Memory Arena Library for Golang
[![Go Report Card](https://goreportcard.com/badge/github.com/Raezil/memoryArena)](https://goreportcard.com/report/github.com/Raezil/memoryArena)

Memory Arena Library is a Golang package that consolidates multiple related memory allocations into a single area. This design allows you to free all allocations at once, making memory management simpler and more efficient.

## Features
- **Generic**: works with any Go type (`int`, structs, pointers, etc.).
- **Grouped Memory Allocations:** Manage related objects within a single arena, streamlining your memory organization.
- **Efficient Cleanup:** Release all allocations in one swift operation, simplifying resource management.
- **Concurrency Support:** Use with concurrent operations via a dedicated concurrent arena.
- **AtomicArena** is a concurrent bump allocator for type-homogeneous objects in Go. It allows safe, lock-free allocations from multiple goroutines using atomic operations, making it well-suited for high-performance, multi-threaded environments.


## Installation

Install the latest version with:

```bash
go get github.com/Raezil/memoryArena@latest
```

## Usage Example

Below is an example demonstrating how to create a memory arena, allocate objects, and free them efficiently:

### Using Memory Arena

```
package main

import (
	"fmt"
	"unsafe"

	. "github.com/Raezil/memoryArena"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	// Allocate enough space for 10 Person structs
	arena, err := NewMemoryArena[Person](10 * int(unsafe.Sizeof(Person{})))
	if err != nil {
		panic(err)
	}
	defer arena.Reset()

	p1, _ := arena.NewObject(Person{"Alice", 30})
	p2, _ := arena.NewObject(Person{"Bob", 25})

	fmt.Println(*p1, *p2)
}
```


### Using Concurrent Arena

```
package main

import (
	"fmt"

	. "github.com/Raezil/memoryArena"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	arena, err := NewConcurrentArena[[]Person](100)
	if err != nil {
		return
	}
	obj, _ := arena.NewObject([]Person{Person{"Kamil", 27}, Person{"Lukasz", 28}})
	defer arena.Reset()
	fmt.Println(obj)

}
```
### Using AtomicArena
```
package main

import (
	"fmt"
	"sync"

	. "github.com/Raezil/memoryArena"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	// Allocate 1KB buffer for Person allocations
	arena, err := NewAtomicArena[Person](1024)
	if err != nil {
		panic(err)
	}
	defer arena.Reset()

	var wg sync.WaitGroup
	for i, name := range []string{"Carol", "Dave", "Eve"} {
		wg.Add(1)
		go func(name string, age int) {
			defer wg.Done()
			p, _ := arena.NewObject(Person{name, age})
			fmt.Println(*p)
		}(name, i+20)
	}
	wg.Wait()
}
```

## Testing & Benchmarks

Run all tests with race detection:

```bash
go test -v -race ./...
```

Run benchmarks:

```bash
go test -bench=. -benchmem
```

Tiny objects (< 32 KB) → stick with new.

Medium to huge buffers (≥ 64 KB) → use a memory/atomic arena to keep latency in single‐digit nanoseconds and avoid GC pressure.

At **100 MB** (100 000 000 B) allocations:

| Strategy               | Latency (ns/op) | Speedup vs `new`      |
|------------------------|-----------------|-----------------------|
| **Native `new`**       | 19 861 997      | 1× (baseline)         |
| **AtomicArena.NewObject** | 7.149           | ~2.8 × 10⁶            |
| **MemoryArena.NewObject** | 4.318           | ~4.6 × 10⁶            |

AtomicArena reduces allocation time by ~2.8 million×.

MemoryArena reduces allocation time by ~4.6 million×.


## **📜 Contributing**
Want to improve memoryArena? 🚀  
1. Fork the repo  
2. Create a feature branch (`git checkout -b feature-new`)  
3. Commit your changes (`git commit -m "Added feature"`)  
4. Push to your branch (`git push origin feature-new`)  
5. Submit a PR!  


