<p align="center">
  <img src="https://github.com/user-attachments/assets/c7f6f25b-e0ce-4159-be8e-7865c6e63236">
</p>

# Memory Arena Library for Golang

Memory Arena Library is a Golang package that consolidates multiple related memory allocations into a single area. This design allows you to free all allocations at once, making memory management simpler and more efficient.

## Features

- **Grouped Memory Allocations:** Allocate related objects in one arena.
- **Efficient Cleanup:** Free all allocations in a single operation.
- **Concurrency Support:** Use with concurrent operations via a dedicated concurrent arena.



## Installation

Install the latest version with:

```bash
go get github.com/Raezil/memoryArena@latest
```

Usage Example

Below is an example demonstrating how to create a memory arena, allocate objects, and free them efficiently:

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
	obj, _ := NewObject[[]Person](arena, []Person{Person{"Kamil", 27}, Person{"Lukasz", 28}})
	defer Reset(arena)
	fmt.Println(obj)

}
```
Testing

To run the tests, execute:
```
go test
```

