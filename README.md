<p align="center">
  <img src="https://github.com/user-attachments/assets/c7f6f25b-e0ce-4159-be8e-7865c6e63236">
</p>

# Memory Arena Library for Golang

Memory Arena Library is a Golang package that consolidates multiple related memory allocations into a single area. This design allows you to free all allocations at once, making memory management simpler and more efficient.

## Features

- **Grouped Memory Allocations:** Manage related objects within a single arena, streamlining your memory organization.
- **Efficient Cleanup:** Release all allocations in one swift operation, simplifying resource management.
- **Concurrency Support:** Use with concurrent operations via a dedicated concurrent arena.



## Installation

Install the latest version with:

```bash
go get github.com/Raezil/memoryArena@latest
```

# Usage Example

Below is an example demonstrating how to create a memory arena, allocate objects, and free them efficiently:

```
package main

import (
	"fmt"
	"log"

	"github.com/Raezil/memoryArena"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	// Create a new concurrent arena for a slice of Person.
	arena, err := memoryArena.NewConcurrentArena[[]Person](100)
	if err != nil {
		log.Fatalf("failed to create arena: %v", err)
	}
	// Ensure the arena is reset at the end.
	defer arena.Reset()

	// Define a slice of Person.
	persons := []Person{
		{"Kamil", 27},
		{"Lukasz", 28},
	}

	// Allocate the slice in the arena.
	allocatedPersons, err := arena.AllocateNewValue(persons)
	if err != nil {
		log.Fatalf("failed to allocate persons: %v", err)
	}

	// Print the allocated slice.
	fmt.Printf("Allocated persons: %+v\n", *allocatedPersons)
}
```
# Testing

To run the tests, execute:
```
go test
```

