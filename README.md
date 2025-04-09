<p align="center">
  <img src="https://github.com/user-attachments/assets/c7f6f25b-e0ce-4159-be8e-7865c6e63236">
</p>

# Memory Arena Library for Golang
[![Go Report Card](https://goreportcard.com/badge/github.com/Raezil/memoryArena)](https://goreportcard.com/report/github.com/Raezil/memoryArena)
[![Coverage](https://img.shields.io/badge/coverage-90.7%25-brightgreen.svg)](https://your-ci-or-coverage-report-url)

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


## **📜 Contributing**
Want to improve memoryArena? 🚀  
1. Fork the repo  
2. Create a feature branch (`git checkout -b feature-new`)  
3. Commit your changes (`git commit -m "Added feature"`)  
4. Push to your branch (`git push origin feature-new`)  
5. Submit a PR!  


